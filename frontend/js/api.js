const API = {
  base: '/api',

  token() { return localStorage.getItem('token'); },

  headers(extra = {}) {
    const h = { 'Content-Type': 'application/json', ...extra };
    const tok = this.token();
    if (tok) h['Authorization'] = 'Bearer ' + tok;
    return h;
  },

  async request(method, path, body) {
    const opts = { method, headers: this.headers() };
    if (body) opts.body = JSON.stringify(body);
    const res = await fetch(this.base + path, opts);
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
    return data;
  },

  async upload(path, formData) {
    const res = await fetch(this.base + path, {
      method: 'POST',
      headers: { 'Authorization': 'Bearer ' + this.token() },
      body: formData,
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
    return data;
  },

  get:    (path)        => API.request('GET',    path),
  post:   (path, body)  => API.request('POST',   path, body),
  put:    (path, body)  => API.request('PUT',    path, body),
  delete: (path)        => API.request('DELETE', path),

  // Auth
  login:  (email, pass) => API.post('/auth/login', { email, password: pass }),
  me:     ()            => API.get('/auth/me'),
  logout: ()            => API.post('/auth/logout'),

  // Cuentas
  cuentas:       ()     => API.get('/cuentas'),
  createCuenta:  (b)    => API.post('/cuentas', b),
  deleteCuenta:  (id)   => API.delete(`/cuentas/${id}`),

  // Cartolas
  cartolas:      ()     => API.get('/cartolas'),
  movimientos:   (id)   => API.get(`/cartolas/${id}/movimientos`),
  importCartola: (fd)   => API.upload('/cartolas/importar', fd),

  // Conciliación
  ejecutarConc:  (id)   => API.post(`/conciliacion/${id}/auto`),
  diferencias:   (id)   => API.get(`/conciliacion/${id}/diferencias`),
  resumenConc:   (id)   => API.get(`/conciliacion/${id}/resumen`),
  resolverDif:   (id, obs) => API.put(`/conciliacion/item/${id}/resolver`, { observacion: obs }),

  // Entidades
  entidades:     ()     => API.get('/entidades'),
  createEntidad: (b)    => API.post('/entidades', b),
  deleteEntidad: (id)   => API.delete(`/entidades/${id}`),

  // Reportes
  posicion:      (d, h) => API.get(`/reportes/posicion?desde=${d}&hasta=${h}`),
  auditoria:     (d, h) => API.get(`/reportes/auditoria?desde=${d}&hasta=${h}`),

  // Perfil
  perfil:        ()     => API.get('/config/perfil'),

  // Jurídico
  custodia:      ()     => API.get('/juridico/custodia'),
  createCustodia:(b)    => API.post('/juridico/custodia', b),
  honorarios:    ()     => API.get('/juridico/honorarios'),
  resumenHon:    ()     => API.get('/juridico/honorarios/resumen'),
  createHon:     (b)    => API.post('/juridico/honorarios', b),

  // Empresa
  categorias:    ()     => API.get('/empresa/categorias'),
  ingresos:      ()     => API.get('/empresa/ingresos'),
  resumenIng:    ()     => API.get('/empresa/ingresos/resumen'),
  createIngreso: (b)    => API.post('/empresa/ingresos', b),
};
