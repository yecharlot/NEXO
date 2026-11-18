package httpapi
import (
 "net/http"
 "sort"
 "strings"
 "time"
 "github.com/yecharlot/NEXO/internal/domain"
)
const eventExtractorVersion="lexical-0.1.0"
var eventSignals=[]string{"comienza","comenzó","comenzo","inicia","inició","inicio","empieza","empezó","empezo","llega","llegó","llego","entra","entró","entro","sale","salió","salio","anuncia","anunció","anuncio","presenta","presentó","presento","descubre","descubrió","descubrio","ocurre","ocurrió","ocurrio","sucede","sucedió","sucedio","termina","terminó","termino","finaliza","finalizó","finalizo","se reúne","se reunió","firma","firmó","firmo","aprueba","aprobó","aprobo","rechaza","rechazó","rechazo","explota","explotó","exploto","muere","murió","murio","nace","ganó","gano","pierde","perdió","perdio","begins","began","starts","started","arrives","arrived","enters","entered","announces","announced","presents","presented","discovers","discovered","happens","occurs","ends","ended","meets","met","signs","signed","approves","approved","rejects","rejected","wins","won","loses","lost","dies","died"}
func (s *Server) events(w http.ResponseWriter,r *http.Request,projectID,resource string){
 db:=s.db.Snapshot();if _,ok:=db.Projects[projectID];!ok{jsonErr(w,404,"project not found");return}
 if resource==""&&r.Method==http.MethodGet{out:=[]domain.Event{};for _,v:=range db.Events{if v.ProjectID==projectID{out=append(out,v)}};sort.Slice(out,func(i,j int)bool{return out[i].CreatedAt.Before(out[j].CreatedAt)});jsonOut(w,200,map[string]any{"project_id":projectID,"events":out});return}
 if resource=="extract"&&r.Method==http.MethodPost{var in struct{};if err:=readJSON(r,&in);err!=nil{jsonErr(w,400,"invalid extraction request");return}
 existing:=map[string]bool{};for _,v:=range db.Events{if v.ProjectID==projectID{for _,eid:=range v.EvidenceIDs{existing[eid]=true}}}
 created,skipped:=0,0;for _,asset:=range db.Assets{if asset.ProjectID!=projectID{continue};for _,seg:=range asset.Segments{if created>=100{skipped++;continue};if existing[seg.ID]||!hasEventSignal(seg.Text){skipped++;continue}
 title:=strings.TrimSpace(seg.Text);runes:=[]rune(title);if len(runes)>180{title=string(runes[:180])+"…"}
 ev:=domain.Event{ID:id("evt"),ProjectID:projectID,AssetID:asset.ID,Title:title,Description:"Propuesta heurística: revisar el contexto antes de aceptarla.",StartSeconds:seg.Start,EndSeconds:seg.End,TimestampSource:seg.TimestampSource,EvidenceIDs:[]string{seg.ID},Status:"pending",Origin:"transcript_rule",Extractor:"lexical-event-signals",ExtractorVersion:eventExtractorVersion,CreatedAt:time.Now().UTC()}
 if err:=s.db.PutEvent(ev);err!=nil{jsonErr(w,500,"cannot persist event proposal");return};existing[seg.ID]=true;created++}}
 jsonOut(w,200,map[string]any{"created":created,"skipped":skipped,"extractor":"lexical-event-signals","version":eventExtractorVersion});return}
 parts:=strings.Split(resource,"/");if len(parts)==2&&parts[1]=="review"&&r.Method==http.MethodPost{ev,ok:=db.Events[parts[0]];if !ok||ev.ProjectID!=projectID{jsonErr(w,404,"event not found");return}
 var in struct{Decision string `json:"decision"`;ReviewNote string `json:"review_note"`;Title string `json:"title"`;Description string `json:"description"`};if err:=readJSON(r,&in);err!=nil{jsonErr(w,400,"invalid review request");return}
 if ev.Status!="pending"{jsonErr(w,409,"event has already been reviewed");return};if in.Decision!="accept"&&in.Decision!="reject"{jsonErr(w,400,"decision must be accept or reject");return};if len(in.ReviewNote)>2000||len(in.Title)>180||len(in.Description)>2000{jsonErr(w,400,"review fields exceed limits");return}
 if strings.TrimSpace(in.Title)!=""{ev.Title=strings.TrimSpace(in.Title)};if strings.TrimSpace(in.Description)!=""{ev.Description=strings.TrimSpace(in.Description)};if in.Decision=="accept"{ev.Status="accepted"}else{ev.Status="rejected"};ev.ReviewedAt=time.Now().UTC();ev.ReviewedBy="human";ev.ReviewNote=strings.TrimSpace(in.ReviewNote)
 if err:=s.db.PutEvent(ev);err!=nil{jsonErr(w,500,"cannot persist event review");return};jsonOut(w,200,ev);return}
 jsonErr(w,404,"event route not found")
}
func hasEventSignal(text string)bool{normalized:=" "+strings.ToLower(strings.Join(strings.Fields(text)," "))+" ";for _,signal:=range eventSignals{if strings.Contains(normalized," "+signal+" ")||strings.HasPrefix(normalized,signal+" ")||strings.HasSuffix(normalized," "+signal){return true}};return false}
