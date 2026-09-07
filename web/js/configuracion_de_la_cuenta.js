(function(){
  var emailEl = document.getElementById('fldEmail');
  var nameEl = document.getElementById('fldName');
  var phoneEl = document.getElementById('fldPhone');
  var saveBtn = document.getElementById('saveProfileBtn');
  var profileMsg = document.getElementById('profileMsg');
  var currentPwd = document.getElementById('fldCurrentPwd');
  var newPwd = document.getElementById('fldNewPwd');
  var newPwdConfirm = document.getElementById('fldNewPwdConfirm');
  var changePwdBtn = document.getElementById('changePwdBtn');
  var pwdMsg = document.getElementById('pwdMsg');

  function show(el, msg, isErr){ if(!el) return; el.style.display='block'; el.textContent = msg; el.style.color = isErr ? '#c0392b' : '#2e7d32'; setTimeout(function(){ el.style.display='none'; },8000); }

  function safeJson(res){ return res.text().then(function(t){ try{ return JSON.parse(t); }catch(e){ return null; }}); }

  function fetchAccount(){ return fetch('/api/account', { credentials: 'same-origin' }).then(function(res){ if(!res.ok) throw new Error('no-auth'); return res.json(); }); }

  fetchAccount().then(function(data){ if(!data) return; if(data.email) emailEl.value = data.email; if(data.name) nameEl.value = data.name; if(data.admin && data.admin.telefono) phoneEl.value = data.admin.telefono || ''; }).catch(function(){ /* ignore */ });

  if (saveBtn) saveBtn.addEventListener('click', function(e){ e.preventDefault(); var payload = { email: (emailEl.value||'').trim(), name: (nameEl.value||'').trim(), telefono: (phoneEl.value||'').trim() }; fetch('/api/account/update_profile', { method:'POST', credentials:'same-origin', headers:{'Content-Type':'application/json'}, body: JSON.stringify(payload) }).then(function(res){ return safeJson(res).then(function(json){ if(res.ok){ show(profileMsg, 'Perfil actualizado', false); } else { show(profileMsg, (json && json.error) ? json.error : 'Error al actualizar', true); } }); }).catch(function(){ show(profileMsg, 'Error al contactar el servidor', true); }); });

  if (changePwdBtn) changePwdBtn.addEventListener('click', function(e){ e.preventDefault(); var cur = (currentPwd.value||'').trim(); var np = (newPwd.value||'').trim(); var np2 = (newPwdConfirm.value||'').trim(); if(!cur || !np){ show(pwdMsg, 'Debe indicar contraseña actual y nueva', true); return; } if(np !== np2){ show(pwdMsg, 'La confirmación no coincide', true); return; } fetch('/api/account/change_password', { method:'POST', credentials:'same-origin', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ current_password: cur, new_password: np }) }).then(function(res){ return safeJson(res).then(function(json){ if(res.ok){ show(pwdMsg, 'Contraseña actualizada', false); currentPwd.value=''; newPwd.value=''; newPwdConfirm.value=''; } else { show(pwdMsg, (json && json.error) ? json.error : 'Error al cambiar contraseña', true); } }); }).catch(function(){ show(pwdMsg, 'Error al contactar servidor', true); }); });
})();
