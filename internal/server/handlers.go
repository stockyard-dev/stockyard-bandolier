package server
import("encoding/json";"net/http";"github.com/stockyard-dev/stockyard-bandolier/internal/store")
func(s *Server)handleListEnvs(w http.ResponseWriter,r *http.Request){list,_:=s.db.ListEnvs();if list==nil{list=[]string{}};writeJSON(w,200,list)}
func(s *Server)handleList(w http.ResponseWriter,r *http.Request){env:=r.PathValue("env");list,_:=s.db.List(env);if list==nil{list=[]store.EnvVar{}};writeJSON(w,200,list)}
func(s *Server)handleSet(w http.ResponseWriter,r *http.Request){env:=r.PathValue("env");var req struct{Key string `json:"key"`;Value string `json:"value"`;ChangedBy string `json:"changed_by"`};json.NewDecoder(r.Body).Decode(&req);if req.Key==""{writeError(w,400,"key required");return};s.db.Set(env,req.Key,req.Value,req.ChangedBy);writeJSON(w,200,map[string]string{"status":"set"})}
func(s *Server)handleHistory(w http.ResponseWriter,r *http.Request){env:=r.PathValue("env");key:=r.PathValue("key");list,_:=s.db.History(env,key);if list==nil{list=[]store.Change{}};writeJSON(w,200,list)}
func(s *Server)handleDelete(w http.ResponseWriter,r *http.Request){env:=r.PathValue("env");key:=r.PathValue("key");s.db.Delete(env,key);writeJSON(w,200,map[string]string{"status":"deleted"})}
func(s *Server)handleOverview(w http.ResponseWriter,r *http.Request){m,_:=s.db.Stats();writeJSON(w,200,m)}
