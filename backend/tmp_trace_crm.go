package main
import (
  "bytes"; "database/sql"; "encoding/json"; "fmt"; "log"; "net/http"; "net/http/httptest"; "net/url"; "os"; "path/filepath"; "strings"; "time"
  dbpkg "github.com/you/pos-backend/db"; "github.com/you/pos-backend/handlers")
func parseEnvFile(path string)(map[string]string,error){out:=map[string]string{};data,err:=os.ReadFile(path);if err!=nil{return nil,err};for _,line:=range strings.Split(string(data),"\n"){line=strings.TrimSpace(line);if line==""||strings.HasPrefix(line,"#"){continue};p:=strings.SplitN(line,"=",2);if len(p)==2{out[strings.TrimSpace(p[0])]=strings.TrimSpace(p[1])}};return out,nil}
func withTunnelPort(dsn,port string)string{if strings.TrimSpace(dsn)==""||strings.TrimSpace(port)==""{return dsn};u,err:=url.Parse(dsn);if err!=nil{return dsn};host:=u.Hostname();if host==""{return dsn};u.Host=host+":"+port;return u.String()}
func openDB(dsn string)(*sql.DB,error){db,err:=sql.Open(dbpkg.PostgresCompatDriverName(),dsn);if err!=nil{return nil,err};if err:=db.Ping();err!=nil{return nil,err};if err:=dbpkg.EnsurePostgresRuntimeCompat(db);err!=nil{return nil,err};return db,nil}
func doJSON(h http.HandlerFunc,m,t string,p any,hd map[string]string)(int,map[string]any,string){var body bytes.Buffer;if p!=nil{_ = json.NewEncoder(&body).Encode(p)};req:=httptest.NewRequest(m,t,&body);if p!=nil{req.Header.Set("Content-Type","application/json")};for k,v:=range hd{req.Header.Set(k,v)};rr:=httptest.NewRecorder();h(rr,req);raw:=rr.Body.String();parsed:=map[string]any{};_ = json.Unmarshal([]byte(raw),&parsed);return rr.Result().StatusCode,parsed,raw}
func main(){_ = os.Setenv("RECAPTCHA_DEV_BYPASS","1"); envs,_:=parseEnvFile(filepath.Join(".env.local")); emp:=envs["DB_EMPRESAS_DSN"]; sup:=envs["DB_SUPERADMIN_DSN"]; if envs["DB_VPS_TUNNEL_ENABLED"]=="1"{port:=strings.TrimSpace(envs["DB_VPS_LOCAL_PORT"]); emp=withTunnelPort(emp,port); sup=withTunnelPort(sup,port)}; dbEmp,_:=openDB(emp); defer dbEmp.Close(); dbSuper,_:=openDB(sup); defer dbSuper.Close(); const empresaID=7; hdr:=map[string]string{"X-Admin-Email":"powerfulcontrolsystem@gmail.com"}; stamp:=time.Now().Format("20060102-150405");
log.Println("setup handlers")
leads:=handlers.WithEmpresaClientesPermissions(dbEmp,dbSuper,handlers.EmpresaCRMLeadsHandler(dbEmp)); inter:=handlers.WithEmpresaClientesPermissions(dbEmp,dbSuper,handlers.EmpresaCRMInteraccionesHandler(dbEmp)); camp:=handlers.WithEmpresaClientesPermissions(dbEmp,dbSuper,handlers.EmpresaCRMCampanasHandler(dbEmp)); cot:=handlers.WithEmpresaVentasPermissions(dbEmp,dbSuper,handlers.EmpresaVentasCotizacionesHandler(dbEmp));
log.Println("lead create")
st,body,raw:=doJSON(leads,"POST",fmt.Sprintf("/api/empresa/crm/leads?empresa_id=%d",empresaID),map[string]any{"empresa_id":empresaID,"nombre":"Lead QA " + stamp,"tipo":"x"},hdr); log.Println("lead",st,raw); leadID:=int64(body["id"].(float64));
log.Println("inter create")
st,body,raw=doJSON(inter,"POST",fmt.Sprintf("/api/empresa/crm/interacciones?empresa_id=%d",empresaID),map[string]any{"empresa_id":empresaID,"lead_id":leadID,"tipo_interaccion":"seguimiento","resumen":"test"},hdr); log.Println("inter",st,raw); _=body
log.Println("camp create")
st,body,raw=doJSON(camp,"POST",fmt.Sprintf("/api/empresa/crm/campanas?empresa_id=%d",empresaID),map[string]any{"empresa_id":empresaID,"nombre":"Camp "+stamp,"canal":"whatsapp"},hdr); log.Println("camp",st,raw); _=body
log.Println("cot create")
st,body,raw=doJSON(cot,"POST",fmt.Sprintf("/api/empresa/ventas/cotizaciones?empresa_id=%d",empresaID),map[string]any{"empresa_id":empresaID,"cliente_nombre":"Lead QA "+stamp,"total":1000},hdr); log.Println("cot create",st,raw); cotID:=int64(body["id"].(float64));
for _,state:= range []string{"emitida","aprobada"}{ log.Println("cot transition",state); st,_,raw=doJSON(cot,"POST",fmt.Sprintf("/api/empresa/ventas/cotizaciones?action=transicionar&empresa_id=%d",empresaID),map[string]any{"empresa_id":empresaID,"id":cotID,"nuevo_estado":state,"estado_documento":state},hdr); log.Println("transition",state,st,raw) }
log.Println("convert pedido")
st,_,raw=doJSON(cot,"POST",fmt.Sprintf("/api/empresa/ventas/cotizaciones?action=convertir_pedido&empresa_id=%d",empresaID),map[string]any{"empresa_id":empresaID,"id":cotID},hdr); log.Println("pedido",st,raw)
log.Println("convert final")
st,_,raw=doJSON(cot,"POST",fmt.Sprintf("/api/empresa/ventas/cotizaciones?action=convertir_documento_final&empresa_id=%d",empresaID),map[string]any{"empresa_id":empresaID,"id":cotID},hdr); log.Println("doc",st,raw)
log.Println("embudo")
st,_,raw=doJSON(cot,"GET",fmt.Sprintf("/api/empresa/ventas/cotizaciones?action=embudo&empresa_id=%d&limit=10",empresaID),nil,hdr); log.Println("embudo",st,raw)
}
