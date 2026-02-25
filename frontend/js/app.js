// ── Estado global ─────────────────────────────────────────────────────────────
const State = {
  user: null,
  perfil: null,
};

// ── Router ────────────────────────────────────────────────────────────────────
const routes = {
  '#dashboard':     viewDashboard,
  '#cuentas':       viewCuentas,
  '#cartolas':      viewCartolas,
  '#conciliacion':  viewConciliacion,
  '#entidades':     viewEntidades,
  '#reportes':      viewReportes,
  '#auditoria':     viewAuditoria,
  '#custodia':      viewCustodia,
  '#honorarios':    viewHonorarios,
  '#ingresos':      viewIngresos,
};

async function navigate() {
  if (!State.user) { renderLogin(); return; }
  const hash = location.hash || '#dashboard';
  const fn = routes[hash];
  if (fn) await fn();
  highlightNav(hash);
}

window.addEventListener('hashchange', navigate);

// ── Bootstrap ─────────────────────────────────────────────────────────────────
(async () => {
  const tok = localStorage.getItem('token');
  if (tok) {
    try {
      State.user = await API.me();
      const cfg = await API.perfil();
      State.perfil = cfg.perfil;
    } catch { localStorage.removeItem('token'); }
  }
  navigate();
})();

// ── Login ─────────────────────────────────────────────────────────────────────
function renderLogin() {
  document.getElementById('app').innerHTML = `
    <div class="login-wrap">
      <div class="login-box">
        <div class="login-title">Conciliación Bancaria</div>
        <div class="login-sub">Ingresa tus credenciales para continuar</div>
        <div id="login-error"></div>
        <div class="form-group">
          <label>Email</label>
          <input id="l-email" type="email" placeholder="admin@conciliacion.local">
        </div>
        <div class="form-group">
          <label>Contraseña</label>
          <input id="l-pass" type="password" placeholder="••••••••">
        </div>
        <button class="btn btn-primary" style="width:100%;margin-top:8px" onclick="doLogin()">
          Ingresar
        </button>
      </div>
    </div>`;
  document.getElementById('l-pass').addEventListener('keydown', e => {
    if (e.key === 'Enter') doLogin();
  });
}

async function doLogin() {
  const email = document.getElementById('l-email').value;
  const pass  = document.getElementById('l-pass').value;
  const err   = document.getElementById('login-error');
  try {
    const res = await API.login(email, pass);
    localStorage.setItem('token', res.token);
    State.user = res.user;
    const cfg = await API.perfil();
    State.perfil = cfg.perfil;
    navigate();
  } catch(e) {
    err.innerHTML = `<div class="alert alert-err">${e.message}</div>`;
  }
}

// ── Shell ─────────────────────────────────────────────────────────────────────
function renderShell(title, bodyHtml) {
  const isJuridico = State.perfil === 'juridico';

  const profileNav = isJuridico ? `
    <div class="nav-section">
      <div class="nav-label">Jurídico</div>
      <a class="nav-item" href="#custodia">
        <span class="nav-icon">⚖</span> Custodia
      </a>
      <a class="nav-item" href="#honorarios">
        <span class="nav-icon">💼</span> Honorarios
      </a>
    </div>` : `
    <div class="nav-section">
      <div class="nav-label">Empresa</div>
      <a class="nav-item" href="#ingresos">
        <span class="nav-icon">↗</span> Pagos
      </a>
    </div>`;

  document.getElementById('app').innerHTML = `
    <div class="layout">
      <aside class="sidebar">
        <div class="sidebar-logo">
          <h1>Conciliación</h1>
          <span>${State.perfil}</span>
        </div>
        <div class="nav-section">
          <div class="nav-label">Principal</div>
          <a class="nav-item" href="#dashboard"><span class="nav-icon">◈</span> Dashboard</a>
          <a class="nav-item" href="#cuentas"><span class="nav-icon">🏦</span> Cuentas</a>
          <a class="nav-item" href="#cartolas"><span class="nav-icon">📄</span> Cartolas</a>
          <a class="nav-item" href="#conciliacion"><span class="nav-icon">⟳</span> Conciliación</a>
          <a class="nav-item" href="#entidades"><span class="nav-icon">👥</span> ${isJuridico ? 'Clientes' : 'Miembros'}</a>
        </div>
        ${profileNav}
        <div class="nav-section">
          <div class="nav-label">Reportes</div>
          <a class="nav-item" href="#reportes"><span class="nav-icon">📊</span> Posición</a>
          ${State.user?.rol === 'ADMIN' ? '<a class="nav-item" href="#auditoria"><span class="nav-icon">🔍</span> Auditoría</a>' : ''}
        </div>
        <div style="margin-top:auto;padding:16px 20px;border-top:1px solid var(--border)">
          <div style="font-size:12px;color:var(--text-muted)">${State.user?.nombre}</div>
          <div style="font-size:11px;color:var(--gold);margin:2px 0">${State.user?.rol}</div>
          <a href="#" onclick="doLogout()" style="font-size:11px;color:var(--text-muted)">Cerrar sesión</a>
        </div>
      </aside>
      <div class="main">
        <div class="topbar">
          <div class="topbar-title">${title}</div>
          <div class="topbar-user">
            <span class="badge-rol">${State.user?.rol}</span>
            ${State.user?.nombre}
          </div>
        </div>
        <div class="content" id="view-content">
          ${bodyHtml}
        </div>
      </div>
    </div>`;
}

function highlightNav(hash) {
  document.querySelectorAll('.nav-item').forEach(el => {
    el.classList.toggle('active', el.getAttribute('href') === hash);
  });
}

async function doLogout() {
  await API.logout().catch(() => {});
  localStorage.removeItem('token');
  State.user = null;
  location.hash = '';
  renderLogin();
}

// ── Dashboard ─────────────────────────────────────────────────────────────────
async function viewDashboard() {
  renderShell('Dashboard', '<div class="empty-state">Cargando...</div>');

  try {
    const hoy = new Date().toISOString().slice(0,10);
    const mesInicio = hoy.slice(0,7) + '-01';
    const pos = await API.posicion(mesInicio, hoy);

    let totalIngresos = 0, totalEgresos = 0;
    pos.posiciones.forEach(p => { totalIngresos += p.ingresos; totalEgresos += p.egresos; });

    let domKpi = '';
    if (State.perfil === 'juridico') {
      const res = await API.resumenHon();
      domKpi = `
        <div class="kpi-cell"><div class="kpi-value">${fmt(res.cobrado)}</div><div class="kpi-label">Honorarios cobrados</div></div>
        <div class="kpi-cell"><div class="kpi-value">${fmt(res.pendiente)}</div><div class="kpi-label">Honorarios pendientes</div></div>`;
    } else {
      const res = await API.resumenIng();
      domKpi = `
        <div class="kpi-cell"><div class="kpi-value">${fmt(res.recibido)}</div><div class="kpi-label">Ingresos recibidos</div></div>
        <div class="kpi-cell"><div class="kpi-value">${fmt(res.pendiente)}</div><div class="kpi-label">Ingresos pendientes</div></div>`;
    }

    document.getElementById('view-content').innerHTML = `
      <div class="kpi-strip">
        <div class="kpi-cell"><div class="kpi-value">${fmt(totalIngresos)}</div><div class="kpi-label">Ingresos del mes</div></div>
        <div class="kpi-cell"><div class="kpi-value">${fmt(totalEgresos)}</div><div class="kpi-label">Egresos del mes</div></div>
        <div class="kpi-cell"><div class="kpi-value">${fmt(totalIngresos - totalEgresos)}</div><div class="kpi-label">Saldo del mes</div></div>
        ${domKpi}
      </div>
      <div class="card">
        <div class="card-header"><div class="card-title">Posición por cuenta — ${mesInicio} al ${hoy}</div></div>
        <div class="table-wrap">
          <table>
            <thead><tr><th>Cuenta</th><th>Banco</th><th class="text-right">Ingresos</th><th class="text-right">Egresos</th><th class="text-right">Saldo</th></tr></thead>
            <tbody>${pos.posiciones.map(p => `
              <tr>
                <td>${p.cuenta || '—'}</td>
                <td>${p.banco}</td>
                <td class="text-right text-gold">${fmt(p.ingresos)}</td>
                <td class="text-right">${fmt(p.egresos)}</td>
                <td class="text-right" style="color:${p.saldo>=0?'#5cc48a':'#e05555'}">${fmt(p.saldo)}</td>
              </tr>`).join('')}
            </tbody>
          </table>
        </div>
      </div>`;
  } catch(e) {
    document.getElementById('view-content').innerHTML = `<div class="alert alert-err">${e.message}</div>`;
  }
}

// ── Cuentas ───────────────────────────────────────────────────────────────────
async function viewCuentas() {
  renderShell('Cuentas Bancarias', '<div class="empty-state">Cargando...</div>');
  const cuentas = await API.cuentas();

  document.getElementById('view-content').innerHTML = `
    <div class="toolbar">
      <span>${cuentas.length} cuenta(s) registrada(s)</span>
      <div class="toolbar-right">
        <button class="btn btn-primary btn-sm" onclick="modalCuenta()">+ Nueva cuenta</button>
      </div>
    </div>
    <div class="card">
      <div class="table-wrap">
        <table>
          <thead><tr><th>Alias</th><th>Banco</th><th>Tipo</th><th>Número</th><th>Moneda</th><th>Estado</th><th></th></tr></thead>
          <tbody>${cuentas.length ? cuentas.map(c => `
            <tr>
              <td>${c.alias || '—'}</td>
              <td>${c.banco}</td>
              <td>${c.tipo}</td>
              <td>${c.numero_cuenta}</td>
              <td>${c.moneda}</td>
              <td><span class="status ${c.activa?'status-ok':'status-err'}">${c.activa?'Activa':'Inactiva'}</span></td>
              <td><button class="btn btn-danger btn-sm" onclick="delCuenta(${c.id})">Eliminar</button></td>
            </tr>`).join('') : '<tr><td colspan="7" class="empty-state">Sin cuentas registradas</td></tr>'}
          </tbody>
        </table>
      </div>
    </div>
    <div id="modal-area"></div>`;
}

function modalCuenta() {
  document.getElementById('modal-area').innerHTML = `
    <div class="modal-overlay" onclick="if(event.target===this)this.remove()">
      <div class="modal">
        <div class="modal-title">Nueva cuenta bancaria</div>
        <div class="form-row">
          <div class="form-group"><label>Banco</label><input id="c-banco" placeholder="BancoEstado"></div>
          <div class="form-group"><label>Número</label><input id="c-num" placeholder="000-000000-00"></div>
        </div>
        <div class="form-row">
          <div class="form-group"><label>Tipo</label>
            <select id="c-tipo"><option>corriente</option><option>vista</option><option>ahorro</option></select>
          </div>
          <div class="form-group"><label>Alias</label><input id="c-alias" placeholder="Cuenta principal"></div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" onclick="this.closest('.modal-overlay').remove()">Cancelar</button>
          <button class="btn btn-primary" onclick="saveCuenta()">Guardar</button>
        </div>
      </div>
    </div>`;
}

async function saveCuenta() {
  try {
    await API.createCuenta({
      banco: document.getElementById('c-banco').value,
      numero_cuenta: document.getElementById('c-num').value,
      tipo: document.getElementById('c-tipo').value,
      alias: document.getElementById('c-alias').value,
      moneda: 'CLP',
    });
    viewCuentas();
  } catch(e) { alert(e.message); }
}

async function delCuenta(id) {
  if (!confirm('¿Eliminar esta cuenta?')) return;
  try { await API.deleteCuenta(id); viewCuentas(); } catch(e) { alert(e.message); }
}

// ── Cartolas ──────────────────────────────────────────────────────────────────
async function viewCartolas() {
  renderShell('Cartolas', '<div class="empty-state">Cargando...</div>');
  const [cartolas, cuentas] = await Promise.all([API.cartolas(), API.cuentas()]);

  const cuentaMap = {};
  cuentas.forEach(c => cuentaMap[c.id] = c.alias || c.banco);

  document.getElementById('view-content').innerHTML = `
    <div class="card">
      <div class="card-header"><div class="card-title">Importar cartola</div></div>
      <div class="form-row">
        <div class="form-group">
          <label>Cuenta bancaria</label>
          <select id="imp-cuenta">
            ${cuentas.map(c => `<option value="${c.id}">${c.alias || c.banco} — ${c.numero_cuenta}</option>`).join('')}
          </select>
        </div>
        <div class="form-group"><label>Período desde</label><input id="imp-desde" type="date"></div>
        <div class="form-group"><label>Período hasta</label><input id="imp-hasta" type="date"></div>
        <div class="form-group">
          <label>Archivo (CSV / TXT / XLSX)</label>
          <input id="imp-file" type="file" accept=".csv,.txt,.xlsx">
        </div>
      </div>
      <button class="btn btn-primary" onclick="importarCartola()">Importar</button>
      <div id="imp-result" class="mt-4"></div>
    </div>
    <div class="card">
      <div class="card-header"><div class="card-title">Cartolas importadas</div></div>
      <div class="table-wrap">
        <table>
          <thead><tr><th>#</th><th>Cuenta</th><th>Desde</th><th>Hasta</th><th>Movimientos</th><th>Archivo</th><th></th></tr></thead>
          <tbody>${cartolas.length ? cartolas.map(c => `
            <tr>
              <td>${c.id}</td>
              <td>${cuentaMap[c.cuenta_id] || c.cuenta_id}</td>
              <td>${c.periodo_desde}</td>
              <td>${c.periodo_hasta}</td>
              <td>${c.total_movimientos}</td>
              <td class="text-muted">${c.archivo_nombre}</td>
              <td>
                <button class="btn btn-secondary btn-sm" onclick="verMovimientos(${c.id})">Ver movimientos</button>
              </td>
            </tr>`).join('') : '<tr><td colspan="7" class="empty-state">Sin cartolas importadas</td></tr>'}
          </tbody>
        </table>
      </div>
    </div>
    <div id="mov-area"></div>`;
}

async function importarCartola() {
  const fileInput = document.getElementById('imp-file');
  const result = document.getElementById('imp-result');
  if (!fileInput.files[0]) { alert('Selecciona un archivo'); return; }

  const fd = new FormData();
  fd.append('archivo', fileInput.files[0]);
  fd.append('cuenta_id', document.getElementById('imp-cuenta').value);
  fd.append('periodo_desde', document.getElementById('imp-desde').value);
  fd.append('periodo_hasta', document.getElementById('imp-hasta').value);

  try {
    const res = await API.importCartola(fd);
    result.innerHTML = `<div class="alert alert-ok">Cartola importada: ${res.movimientos} movimientos.</div>`;
    setTimeout(viewCartolas, 1500);
  } catch(e) {
    result.innerHTML = `<div class="alert alert-err">${e.message}</div>`;
  }
}

async function verMovimientos(cartolaId) {
  const movs = await API.movimientos(cartolaId);
  document.getElementById('mov-area').innerHTML = `
    <div class="card">
      <div class="card-header">
        <div class="card-title">Movimientos — Cartola #${cartolaId}</div>
        <button class="btn btn-secondary btn-sm" onclick="this.closest('.card').remove()">Cerrar</button>
      </div>
      <div class="table-wrap">
        <table>
          <thead><tr><th>Fecha</th><th>Descripción</th><th>Referencia</th><th class="text-right">Debe</th><th class="text-right">Haber</th><th class="text-right">Saldo</th><th>Estado</th></tr></thead>
          <tbody>${movs.map(m => `
            <tr>
              <td>${m.fecha}</td>
              <td>${m.descripcion || '—'}</td>
              <td class="text-muted">${m.referencia || '—'}</td>
              <td class="text-right">${m.debe > 0 ? fmt(m.debe) : '—'}</td>
              <td class="text-right text-gold">${m.haber > 0 ? fmt(m.haber) : '—'}</td>
              <td class="text-right">${fmt(m.saldo)}</td>
              <td><span class="status ${m.conciliado?'status-ok':'status-pending'}">${m.conciliado?'Conciliado':'Pendiente'}</span></td>
            </tr>`).join('')}
          </tbody>
        </table>
      </div>
    </div>`;
}

// ── Conciliación ──────────────────────────────────────────────────────────────
async function viewConciliacion() {
  renderShell('Conciliación', '<div class="empty-state">Cargando...</div>');
  const cartolas = await API.cartolas();

  document.getElementById('view-content').innerHTML = `
    <div class="card">
      <div class="card-header"><div class="card-title">Ejecutar conciliación automática</div></div>
      <div class="form-row" style="max-width:400px">
        <div class="form-group">
          <label>Cartola</label>
          <select id="conc-cartola">
            ${cartolas.map(c => `<option value="${c.id}">Cartola #${c.id} — ${c.periodo_desde} → ${c.periodo_hasta} (${c.total_movimientos} mov.)</option>`).join('')}
          </select>
        </div>
      </div>
      <button class="btn btn-primary" onclick="ejecutarConc()">Ejecutar conciliación</button>
      <div id="conc-result" class="mt-4"></div>
    </div>
    <div id="dif-area"></div>`;
}

async function ejecutarConc() {
  const id = document.getElementById('conc-cartola').value;
  const res = document.getElementById('conc-result');
  try {
    const data = await API.ejecutarConc(id);
    const r = data.resumen;
    res.innerHTML = `
      <div class="alert alert-ok">
        Conciliación completada: ${r.conciliados} conciliados · ${r.diferencias} diferencias · ${r.pendientes} pendientes · ${r.pct_conciliado.toFixed(1)}% conciliado
      </div>`;
    if (r.diferencias > 0) verDiferencias(id);
  } catch(e) { res.innerHTML = `<div class="alert alert-err">${e.message}</div>`; }
}

async function verDiferencias(cartolaId) {
  const difs = await API.diferencias(cartolaId);
  document.getElementById('dif-area').innerHTML = `
    <div class="card">
      <div class="card-header"><div class="card-title">Diferencias — Cartola #${cartolaId}</div></div>
      <div class="table-wrap">
        <table>
          <thead><tr><th>#</th><th>Fecha</th><th>Descripción</th><th class="text-right">Debe</th><th class="text-right">Haber</th><th>Observación</th><th></th></tr></thead>
          <tbody>${difs.map(d => `
            <tr>
              <td>${d.id}</td>
              <td>${d.fecha || '—'}</td>
              <td>${d.descripcion || '—'}</td>
              <td class="text-right">${d.debe > 0 ? fmt(d.debe) : '—'}</td>
              <td class="text-right">${d.haber > 0 ? fmt(d.haber) : '—'}</td>
              <td class="text-muted">${d.observacion || '—'}</td>
              <td><button class="btn btn-secondary btn-sm" onclick="resolverDif(${d.id})">Resolver</button></td>
            </tr>`).join('')}
          </tbody>
        </table>
      </div>
    </div>`;
}

async function resolverDif(id) {
  const obs = prompt('Ingresa una justificación para resolver esta diferencia:');
  if (obs === null) return;
  try {
    await API.resolverDif(id, obs);
    document.querySelector(`button[onclick="resolverDif(${id})"]`).closest('tr').remove();
  } catch(e) { alert(e.message); }
}

// ── Entidades ─────────────────────────────────────────────────────────────────
async function viewEntidades() {
  const label = State.perfil === 'juridico' ? 'Clientes' : 'Miembros';
  renderShell(label, '<div class="empty-state">Cargando...</div>');
  const data = await API.entidades();

  document.getElementById('view-content').innerHTML = `
    <div class="toolbar">
      <span>${data.entidades.length} ${label.toLowerCase()}</span>
      <div class="toolbar-right">
        <button class="btn btn-primary btn-sm" onclick="modalEntidad()">+ Nuevo/a</button>
      </div>
    </div>
    <div class="card">
      <div class="table-wrap">
        <table>
          <thead><tr><th>Nombre</th><th>RUT</th><th>Email</th><th>Teléfono</th><th>Estado</th><th></th></tr></thead>
          <tbody>${data.entidades.length ? data.entidades.map(e => `
            <tr>
              <td>${e.nombre}</td>
              <td>${e.rut || '—'}</td>
              <td>${e.email || '—'}</td>
              <td>${e.telefono || '—'}</td>
              <td><span class="status ${e.activo?'status-ok':'status-err'}">${e.activo?'Activo':'Inactivo'}</span></td>
              <td><button class="btn btn-danger btn-sm" onclick="delEntidad(${e.id})">Eliminar</button></td>
            </tr>`).join('') : `<tr><td colspan="6" class="empty-state">Sin ${label.toLowerCase()} registrados</td></tr>`}
          </tbody>
        </table>
      </div>
    </div>
    <div id="modal-area"></div>`;
}

function modalEntidad() {
  const label = State.perfil === 'juridico' ? 'cliente' : 'miembro';
  document.getElementById('modal-area').innerHTML = `
    <div class="modal-overlay" onclick="if(event.target===this)this.remove()">
      <div class="modal">
        <div class="modal-title">Nuevo/a ${label}</div>
        <div class="form-group"><label>Nombre</label><input id="e-nombre"></div>
        <div class="form-row">
          <div class="form-group"><label>RUT</label><input id="e-rut" placeholder="12.345.678-9"></div>
          <div class="form-group"><label>Email</label><input id="e-email" type="email"></div>
        </div>
        <div class="form-group"><label>Teléfono</label><input id="e-tel" placeholder="+56 9 XXXX XXXX"></div>
        <div class="modal-footer">
          <button class="btn btn-secondary" onclick="this.closest('.modal-overlay').remove()">Cancelar</button>
          <button class="btn btn-primary" onclick="saveEntidad()">Guardar</button>
        </div>
      </div>
    </div>`;
}

async function saveEntidad() {
  try {
    await API.createEntidad({
      nombre: document.getElementById('e-nombre').value,
      rut: document.getElementById('e-rut').value,
      email: document.getElementById('e-email').value,
      telefono: document.getElementById('e-tel').value,
    });
    viewEntidades();
  } catch(e) { alert(e.message); }
}

async function delEntidad(id) {
  if (!confirm('¿Eliminar?')) return;
  try { await API.deleteEntidad(id); viewEntidades(); } catch(e) { alert(e.message); }
}

// ── Reportes ──────────────────────────────────────────────────────────────────
async function viewReportes() {
  renderShell('Posición Financiera', `
    <div class="card">
      <div class="card-header"><div class="card-title">Filtros</div></div>
      <div class="form-row" style="max-width:500px">
        <div class="form-group"><label>Desde</label><input id="r-desde" type="date"></div>
        <div class="form-group"><label>Hasta</label><input id="r-hasta" type="date"></div>
      </div>
      <button class="btn btn-primary" onclick="cargarPosicion()">Consultar</button>
    </div>
    <div id="pos-result"></div>`);

  // Defaults: mes actual
  const hoy = new Date().toISOString().slice(0,10);
  document.getElementById('r-desde').value = hoy.slice(0,7) + '-01';
  document.getElementById('r-hasta').value = hoy;
}

async function cargarPosicion() {
  const desde = document.getElementById('r-desde').value;
  const hasta = document.getElementById('r-hasta').value;
  try {
    const data = await API.posicion(desde, hasta);
    document.getElementById('pos-result').innerHTML = `
      <div class="card">
        <div class="card-header"><div class="card-title">Posición ${desde} → ${hasta}</div></div>
        <div class="table-wrap">
          <table>
            <thead><tr><th>Cuenta</th><th>Banco</th><th class="text-right">Ingresos</th><th class="text-right">Egresos</th><th class="text-right">Saldo</th></tr></thead>
            <tbody>${data.posiciones.map(p => `
              <tr>
                <td>${p.cuenta || '—'}</td>
                <td>${p.banco}</td>
                <td class="text-right text-gold">${fmt(p.ingresos)}</td>
                <td class="text-right">${fmt(p.egresos)}</td>
                <td class="text-right" style="color:${p.saldo>=0?'#5cc48a':'#e05555'}">${fmt(p.saldo)}</td>
              </tr>`).join('')}
            </tbody>
          </table>
        </div>
      </div>`;
  } catch(e) { alert(e.message); }
}

// ── Auditoría ─────────────────────────────────────────────────────────────────
async function viewAuditoria() {
  renderShell('Auditoría', '<div class="empty-state">Cargando...</div>');
  const hoy = new Date().toISOString().slice(0,10);
  const data = await API.auditoria('2000-01-01', hoy);
  document.getElementById('view-content').innerHTML = `
    <div class="card">
      <div class="card-header"><div class="card-title">Log de operaciones (últimas 200)</div></div>
      <div class="table-wrap">
        <table>
          <thead><tr><th>Fecha</th><th>Usuario</th><th>Acción</th><th>Tabla</th><th>ID</th><th>IP</th></tr></thead>
          <tbody>${data.map(e => `
            <tr>
              <td style="white-space:nowrap">${new Date(e.created_at).toLocaleString('es-CL')}</td>
              <td>${e.usuario_id || '—'}</td>
              <td><span class="status ${e.accion==='DELETE'?'status-err':e.accion==='CREATE'?'status-ok':'status-warn'}">${e.accion}</span></td>
              <td>${e.tabla}</td>
              <td>${e.registro_id}</td>
              <td class="text-muted">${e.ip || '—'}</td>
            </tr>`).join('')}
          </tbody>
        </table>
      </div>
    </div>`;
}

// ── Custodia (jurídico) ───────────────────────────────────────────────────────
async function viewCustodia() {
  renderShell('Fondos en Custodia', '<div class="empty-state">Cargando...</div>');
  const [items, entidades] = await Promise.all([API.custodia(), API.entidades()]);
  const total = items.reduce((s, c) => s + (c.estado === 'activo' ? c.monto : 0), 0);

  document.getElementById('view-content').innerHTML = `
    <div class="kpi-strip" style="margin-bottom:24px">
      <div class="kpi-cell"><div class="kpi-value">${fmt(total)}</div><div class="kpi-label">Total en custodia activa</div></div>
    </div>
    <div class="toolbar">
      <span>${items.length} registros</span>
      <button class="btn btn-primary btn-sm" onclick="modalCustodia()">+ Registrar custodia</button>
    </div>
    <div class="card">
      <div class="table-wrap">
        <table>
          <thead><tr><th>Cliente</th><th>Descripción</th><th class="text-right">Monto</th><th>Ingreso</th><th>Devolución</th><th>Estado</th></tr></thead>
          <tbody>${items.map(c => `
            <tr>
              <td>${c.entidad_nombre}</td>
              <td>${c.descripcion}</td>
              <td class="text-right text-gold">${fmt(c.monto)}</td>
              <td>${c.fecha_ingreso}</td>
              <td>${c.fecha_devolucion || '—'}</td>
              <td><span class="status ${c.estado==='activo'?'status-ok':'status-pending'}">${c.estado}</span></td>
            </tr>`).join('')}
          </tbody>
        </table>
      </div>
    </div>
    <div id="modal-area"></div>`;

  window._entidades = entidades.entidades;
}

function modalCustodia() {
  const opts = (window._entidades||[]).map(e => `<option value="${e.id}">${e.nombre}</option>`).join('');
  document.getElementById('modal-area').innerHTML = `
    <div class="modal-overlay" onclick="if(event.target===this)this.remove()">
      <div class="modal">
        <div class="modal-title">Registrar fondo en custodia</div>
        <div class="form-group"><label>Cliente</label><select id="cu-ent">${opts}</select></div>
        <div class="form-group"><label>Descripción</label><input id="cu-desc" placeholder="Ej: Fondos causa C-123-2026"></div>
        <div class="form-row">
          <div class="form-group"><label>Monto (CLP)</label><input id="cu-monto" type="number" step="1"></div>
          <div class="form-group"><label>Fecha ingreso</label><input id="cu-fecha" type="date"></div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" onclick="this.closest('.modal-overlay').remove()">Cancelar</button>
          <button class="btn btn-primary" onclick="saveCustodia()">Guardar</button>
        </div>
      </div>
    </div>`;
}

async function saveCustodia() {
  try {
    await API.createCustodia({
      entidad_id: parseInt(document.getElementById('cu-ent').value),
      descripcion: document.getElementById('cu-desc').value,
      monto: parseFloat(document.getElementById('cu-monto').value),
      fecha_ingreso: document.getElementById('cu-fecha').value,
    });
    viewCustodia();
  } catch(e) { alert(e.message); }
}

// ── Honorarios (jurídico) ─────────────────────────────────────────────────────
async function viewHonorarios() {
  renderShell('Honorarios', '<div class="empty-state">Cargando...</div>');
  const [items, entidades, resumen] = await Promise.all([API.honorarios(), API.entidades(), API.resumenHon()]);

  document.getElementById('view-content').innerHTML = `
    <div class="kpi-strip" style="margin-bottom:24px">
      <div class="kpi-cell"><div class="kpi-value">${fmt(resumen.pactado)}</div><div class="kpi-label">Pactado total</div></div>
      <div class="kpi-cell"><div class="kpi-value">${fmt(resumen.cobrado)}</div><div class="kpi-label">Cobrado</div></div>
      <div class="kpi-cell"><div class="kpi-value">${fmt(resumen.pendiente)}</div><div class="kpi-label">Pendiente</div></div>
    </div>
    <div class="toolbar">
      <span>${items.length} honorarios</span>
      <button class="btn btn-primary btn-sm" onclick="modalHonorario()">+ Registrar</button>
    </div>
    <div class="card">
      <div class="table-wrap">
        <table>
          <thead><tr><th>Cliente</th><th>Concepto</th><th class="text-right">Monto</th><th>Emisión</th><th>Pago</th><th>Estado</th></tr></thead>
          <tbody>${items.map(h => `
            <tr>
              <td>${h.entidad_nombre}</td>
              <td>${h.concepto}</td>
              <td class="text-right text-gold">${fmt(h.monto)}</td>
              <td>${h.fecha_emision}</td>
              <td>${h.fecha_pago || '—'}</td>
              <td><span class="status ${h.estado==='pagado'?'status-ok':h.estado==='pendiente'?'status-warn':'status-err'}">${h.estado}</span></td>
            </tr>`).join('')}
          </tbody>
        </table>
      </div>
    </div>
    <div id="modal-area"></div>`;

  window._entidades = entidades.entidades;
}

function modalHonorario() {
  const opts = (window._entidades||[]).map(e => `<option value="${e.id}">${e.nombre}</option>`).join('');
  document.getElementById('modal-area').innerHTML = `
    <div class="modal-overlay" onclick="if(event.target===this)this.remove()">
      <div class="modal">
        <div class="modal-title">Registrar honorario</div>
        <div class="form-group"><label>Cliente</label><select id="h-ent">${opts}</select></div>
        <div class="form-group"><label>Concepto</label><input id="h-conc" placeholder="Ej: Honorarios causa C-123-2026"></div>
        <div class="form-row">
          <div class="form-group"><label>Monto (CLP)</label><input id="h-monto" type="number" step="1"></div>
          <div class="form-group"><label>Fecha emisión</label><input id="h-fecha" type="date"></div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" onclick="this.closest('.modal-overlay').remove()">Cancelar</button>
          <button class="btn btn-primary" onclick="saveHonorario()">Guardar</button>
        </div>
      </div>
    </div>`;
}

async function saveHonorario() {
  try {
    await API.createHon({
      entidad_id: parseInt(document.getElementById('h-ent').value),
      concepto: document.getElementById('h-conc').value,
      monto: parseFloat(document.getElementById('h-monto').value),
      fecha_emision: document.getElementById('h-fecha').value,
    });
    viewHonorarios();
  } catch(e) { alert(e.message); }
}

// ── Ingresos / Pagos esperados (empresa) ─────────────────────────────────────
async function viewIngresos() {
  renderShell('Pagos', '<div class="empty-state">Cargando...</div>');
  const [items, resumen, entidades, categorias] = await Promise.all([
    API.ingresos(), API.resumenIng(), API.entidades(), API.categorias()
  ]);

  window._entidades   = entidades.entidades;
  window._categorias  = categorias;

  const unicos      = items.filter(i => i.periodicidad === 'unico');
  const recurrentes = items.filter(i => i.periodicidad !== 'unico');

  const filaIngreso = (i) => `
    <tr>
      <td>${i.entidad_nombre}</td>
      <td>${i.categoria_nombre}</td>
      <td>${i.descripcion}</td>
      <td class="text-right text-gold">${fmt(i.monto)}</td>
      <td><span class="status ${i.estado==='recibido'?'status-ok':'status-warn'}">${i.estado}</span></td>
      <td>
        ${i.estado === 'esperado'
          ? `<button class="btn btn-secondary btn-sm" onclick="marcarRecibido(${i.id})">Marcar recibido</button>`
          : '<span class="text-muted" style="font-size:11px">—</span>'}
      </td>
    </tr>`;

  document.getElementById('view-content').innerHTML = `
    <div class="kpi-strip" style="margin-bottom:24px">
      <div class="kpi-cell"><div class="kpi-value">${fmt(resumen.esperado)}</div><div class="kpi-label">Esperado total</div></div>
      <div class="kpi-cell"><div class="kpi-value">${fmt(resumen.recibido)}</div><div class="kpi-label">Recibido</div></div>
      <div class="kpi-cell"><div class="kpi-value">${fmt(resumen.pendiente)}</div><div class="kpi-label">Pendiente</div></div>
    </div>

    <div class="card">
      <div class="card-header">
        <div class="card-title">Pagos únicos esperados</div>
        <button class="btn btn-primary btn-sm" onclick="modalPago('unico')">+ Registrar pago esperado</button>
      </div>
      <div class="table-wrap">
        <table>
          <thead><tr><th>De quién</th><th>Categoría</th><th>Concepto</th><th class="text-right">Monto</th><th>Estado</th><th></th></tr></thead>
          <tbody>${unicos.length
            ? unicos.map(filaIngreso).join('')
            : '<tr><td colspan="6" class="empty-state">Sin pagos únicos registrados</td></tr>'}
          </tbody>
        </table>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <div class="card-title">Suscripciones recurrentes</div>
        <button class="btn btn-secondary btn-sm" onclick="modalPago('mensual')">+ Agregar suscripción</button>
      </div>
      <div class="table-wrap">
        <table>
          <thead><tr><th>Miembro</th><th>Categoría</th><th>Descripción</th><th class="text-right">Monto</th><th>Estado</th><th></th></tr></thead>
          <tbody>${recurrentes.length
            ? recurrentes.map(filaIngreso).join('')
            : '<tr><td colspan="6" class="empty-state">Sin suscripciones registradas aún</td></tr>'}
          </tbody>
        </table>
      </div>
    </div>
    <div id="modal-area"></div>`;
}

function modalPago(periodoDefault) {
  const esUnico = periodoDefault === 'unico';
  const entOpts = (window._entidades||[]).map(e => `<option value="${e.id}">${e.nombre}</option>`).join('');
  const catOpts = (window._categorias||[]).map(c => `<option value="${c.id}">${c.nombre}</option>`).join('');

  document.getElementById('modal-area').innerHTML = `
    <div class="modal-overlay" onclick="if(event.target===this)this.remove()">
      <div class="modal">
        <div class="modal-title">${esUnico ? 'Registrar pago esperado' : 'Agregar suscripción'}</div>
        <div class="form-group"><label>${esUnico ? 'De quién' : 'Miembro'}</label><select id="p-ent">${entOpts}</select></div>
        <div class="form-group"><label>Categoría</label><select id="p-cat">${catOpts || '<option value="">Sin categorías</option>'}</select></div>
        <div class="form-group"><label>${esUnico ? 'Concepto' : 'Descripción'}</label>
          <input id="p-desc" placeholder="${esUnico ? 'Ej: Pago pendiente factura #123' : 'Ej: Membresía mensual Plan Pro'}">
        </div>
        <div class="form-row">
          <div class="form-group"><label>Monto (CLP)</label><input id="p-monto" type="number" step="1" min="1"></div>
          ${esUnico ? '' : `
          <div class="form-group"><label>Periodicidad</label>
            <select id="p-periodo">
              <option value="mensual">Mensual</option>
              <option value="trimestral">Trimestral</option>
              <option value="anual">Anual</option>
            </select>
          </div>`}
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" onclick="this.closest('.modal-overlay').remove()">Cancelar</button>
          <button class="btn btn-primary" onclick="savePago('${periodoDefault}')">Guardar</button>
        </div>
      </div>
    </div>`;
}

async function savePago(periodoDefault) {
  const esUnico = periodoDefault === 'unico';
  const periodo = esUnico ? 'unico' : (document.getElementById('p-periodo')?.value || 'mensual');
  try {
    await API.createIngreso({
      entidad_id:   parseInt(document.getElementById('p-ent').value),
      categoria_id: parseInt(document.getElementById('p-cat').value),
      descripcion:  document.getElementById('p-desc').value,
      monto:        parseFloat(document.getElementById('p-monto').value),
      periodicidad: periodo,
      estado:       'esperado',
    });
    viewIngresos();
  } catch(e) { alert(e.message); }
}

async function marcarRecibido(id) {
  try {
    await API.put(`/empresa/ingresos/${id}`, { estado: 'recibido', activo: true });
    viewIngresos();
  } catch(e) { alert(e.message); }
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function fmt(n) {
  if (n === null || n === undefined) return '—';
  return new Intl.NumberFormat('es-CL', { style: 'currency', currency: 'CLP', minimumFractionDigits: 0 }).format(n);
}
