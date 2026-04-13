package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type Usuario struct {
	Idade    int    `json:"idade"`
	Nome     string `json:"nome"`
	CriadoEm string `json:"criado_em,omitempty"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "./usuarios.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS usuarios (
		idade INTEGER, nome TEXT NOT NULL UNIQUE, criado_em TEXT NOT NULL
	)`); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlUI))
	})
	http.HandleFunc("/usuarios", rotaUsuarios)
	http.HandleFunc("/usuarios/", rotaUsuarios)

	log.Println("Servidor rodando em http://localhost:6820")
	log.Fatal(http.ListenAndServe(":6820", nil))
}

func rotaUsuarios(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	nome := r.URL.Path[len("/usuarios"):]
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet:
		filtro := r.URL.Query().Get("nome")
		query := "SELECT idade, nome, criado_em FROM usuarios ORDER BY criado_em DESC"
		args := []any{}
		if filtro != "" {
			query = "SELECT idade, nome, criado_em FROM usuarios WHERE nome = ?"
			args = append(args, filtro)
		}
		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		lista := []Usuario{}
		for rows.Next() {
			var u Usuario
			if err := rows.Scan(&u.Idade, &u.Nome, &u.CriadoEm); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			lista = append(lista, u)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(lista)

	case r.Method == http.MethodPost:
		var u Usuario
		if json.NewDecoder(r.Body).Decode(&u) != nil || u.Nome == "" {
			http.Error(w, "JSON inválido ou nome vazio", 400)
			return
		}
		if u.Idade < 0 || u.Idade > 150 {
			http.Error(w, "Idade deve estar entre 0 e 150", 400)
			return
		}
		u.CriadoEm = time.Now().Format("2006-01-02 15:04:05")
		if _, err := db.Exec("INSERT INTO usuarios VALUES (?,?,?)", u.Idade, u.Nome, u.CriadoEm); err != nil {
			http.Error(w, "Usuário já existe: "+u.Nome, 409)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(u)

	case r.Method == http.MethodDelete && len(nome) > 1:
		nome = nome[1:] 
		res, err := db.Exec("DELETE FROM usuarios WHERE nome = ?", nome)
		if err != nil {
			http.Error(w, "Erro ao deletar: "+err.Error(), 500)
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			http.Error(w, "Usuário não encontrado", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"mensagem": "Deletado: " + nome})

	default:
		http.Error(w, "Método não permitido", 405)
	}
}

const htmlUI = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Usuários</title>
<link href="https://fonts.googleapis.com/css2?family=Syne:wght@700;800&family=DM+Mono&display=swap" rel="stylesheet">
<style>
*{box-sizing:border-box;margin:0;padding:0}
:root{--bg:#0d0d0f;--s:#16161a;--b:#2a2a32;--a:#c8f135;--a2:#7b61ff;--t:#e8e8f0;--m:#6b6b80;--d:#ff4d6d}
body{background:var(--bg);color:var(--t);font-family:'Syne',sans-serif;padding:2rem}
header{display:flex;align-items:baseline;gap:1rem;border-bottom:1px solid var(--b);padding-bottom:1.5rem;margin-bottom:2.5rem}
h1{font-size:2rem;font-weight:800;color:var(--a)}
header span{font:0.75rem 'DM Mono',monospace;color:var(--m);background:var(--s);border:1px solid var(--b);padding:2px 8px;border-radius:4px}
.layout{display:grid;grid-template-columns:320px 1fr;gap:2rem;align-items:start}
.card{background:var(--s);border:1px solid var(--b);border-radius:12px;padding:1.5rem}
h2{font:0.65rem 'DM Mono',monospace;letter-spacing:.12em;text-transform:uppercase;color:var(--m);margin-bottom:1.2rem}
label{display:block;font:0.7rem 'DM Mono',monospace;color:var(--m);text-transform:uppercase;margin-bottom:.3rem}
input{width:100%;background:var(--bg);border:1px solid var(--b);border-radius:8px;padding:.6rem .9rem;color:var(--t);font:0.9rem 'DM Mono',monospace;outline:none;margin-bottom:1rem}
input:focus{border-color:var(--a2)}
.row{display:flex;gap:.6rem;margin-bottom:1rem}
.row input{margin:0}
button{background:var(--a);border:none;border-radius:8px;padding:.7rem 1.2rem;font:700 .85rem 'Syne',sans-serif;cursor:pointer;transition:opacity .15s}
button:hover{opacity:.85}
.sb{background:var(--b);color:var(--t);font-size:.8rem;padding:0 1rem}
.sb:hover{background:var(--a2)}
table{width:100%;border-collapse:collapse}
th{font:0.65rem 'DM Mono',monospace;letter-spacing:.1em;text-transform:uppercase;color:var(--m);text-align:left;padding:0 .8rem .8rem;border-bottom:1px solid var(--b)}
td{padding:.7rem .8rem;font:0.85rem 'DM Mono',monospace;border-bottom:1px solid var(--b)}
td.ts{color:var(--m);font-size:.72rem}
.badge{background:rgba(123,97,255,.15);color:var(--a2);border:1px solid rgba(123,97,255,.3);border-radius:4px;padding:2px 8px}
.del{background:none;border:1px solid var(--b);border-radius:6px;color:var(--m);font:0.7rem 'DM Mono',monospace;padding:3px 10px}
.del:hover{border-color:var(--d);color:var(--d);background:none;opacity:1}
.empty{text-align:center;padding:3rem;color:var(--m);font:0.8rem 'DM Mono',monospace}
.toast{position:fixed;bottom:2rem;right:2rem;padding:.8rem 1.4rem;border-radius:10px;font:0.82rem 'DM Mono',monospace;opacity:0;transform:translateY(10px);transition:all .25s;pointer-events:none}
.toast.show{opacity:1;transform:translateY(0)}
.toast.ok{background:var(--a);color:#0d0d0f}
.toast.err{background:var(--d);color:#fff}
@media(max-width:700px){.layout{grid-template-columns:1fr}}
</style>
</head>
<body>
<header><h1>Usuários</h1><span>localhost:6820</span></header>
<div class="layout">
  <div class="card">
    <h2>Novo usuário</h2>
    <label>Nome</label><input id="n" placeholder="ex: Maria">
    <label>Idade</label><input id="i" type="number" placeholder="ex: 28" min="0">
    <button onclick="criar()">Cadastrar</button>
  </div>
  <div class="card">
    <h2>Lista</h2>
    <div class="row">
      <input id="f" placeholder="Filtrar por nome…" onkeydown="if(event.key==='Enter')buscar()">
      <button class="sb" onclick="buscar()">Buscar</button>
      <button class="sb" onclick="listar()">Todos</button>
    </div>
    <table>
      <thead><tr><th>Nome</th><th>Idade</th><th>Adicionado em</th><th></th></tr></thead>
      <tbody id="tb"></tbody>
    </table>
    <div id="empty" class="empty" style="display:none">Nenhum usuário encontrado.</div>
  </div>
</div>
<div class="toast" id="toast"></div>
<script>
const $=id=>document.getElementById(id);
async function listar(q=''){
  const r=await fetch('/usuarios'+(q?'?nome='+encodeURIComponent(q):''));
  const d=await r.json();
  $('tb').innerHTML='';
  $('empty').style.display=d.length?'none':'block';
  d.forEach(u=>{
    const tr=document.createElement('tr');
    const td1=document.createElement('td');td1.textContent=u.nome;
    const td2=document.createElement('td');const badge=document.createElement('span');badge.className='badge';badge.textContent=u.idade;td2.appendChild(badge);
    const td3=document.createElement('td');td3.className='ts';td3.textContent=u.criado_em||'—';
    const td4=document.createElement('td');const btn=document.createElement('button');btn.className='del';btn.textContent='remover';btn.onclick=()=>del(u.nome);td4.appendChild(btn);
    tr.append(td1,td2,td3,td4);
    $('tb').appendChild(tr);
  });
}
async function criar(){
  const nome=$('n').value.trim(), idade=parseInt($('i').value);
  if(!nome||isNaN(idade)){toast('Preencha nome e idade',0);return}
  const r=await fetch('/usuarios',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({nome,idade})});
  if(r.ok){toast('Cadastrado!',1);$('n').value='';$('i').value='';listar()}
  else toast((await r.text()).trim(),0);
}
async function del(nome){
  if(!confirm('Remover "'+nome+'"?'))return;
  const r=await fetch('/usuarios/'+encodeURIComponent(nome),{method:'DELETE'});
  r.ok?toast('Removido!',1):toast('Erro',0);
  listar();
}
function buscar(){listar($('f').value.trim())}
function toast(msg,ok){
  const e=$('toast');
  e.textContent=msg;e.className='toast show '+(ok?'ok':'err');
  setTimeout(()=>e.className='toast',2500);
}
listar();
</script>
</body>
</html>`
