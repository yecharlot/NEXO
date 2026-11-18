package httpapi

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "path/filepath"
 "strings"
 "testing"
 "github.com/yecharlot/NEXO/internal/domain"
 "github.com/yecharlot/NEXO/internal/store"
)
func TestEventExtractionAndHumanReview(t *testing.T){
 db,err:=store.Open(filepath.Join(t.TempDir(),"nexo.json"));if err!=nil{t.Fatal(err)}
 if err=db.PutProject(domain.Project{ID:"p",Name:"Test"});err!=nil{t.Fatal(err)}
 if err=db.PutAsset(domain.Asset{ID:"a",ProjectID:"p",Title:"Fuente",Segments:[]domain.Segment{{ID:"seg-1",Start:12,End:19,Text:"La empresa anuncia el nuevo proyecto.",TimestampSource:"exact"},{ID:"seg-2",Start:20,End:25,Text:"La charla trata sobre energía solar.",TimestampSource:"exact"}}});err!=nil{t.Fatal(err)}
 ts:=httptest.NewServer(New(db,t.TempDir(),"","","","","","","").Handler());defer ts.Close()
 res,err:=http.Post(ts.URL+"/api/projects/p/events/extract","application/json",strings.NewReader("{}"));if err!=nil{t.Fatal(err)}
 var extracted struct{Created int `json:"created"`;Skipped int `json:"skipped"`};if err=json.NewDecoder(res.Body).Decode(&extracted);err!=nil{t.Fatal(err)};res.Body.Close()
 if res.StatusCode!=http.StatusOK||extracted.Created!=1{t.Fatalf("unexpected extraction result: status=%d body=%+v",res.StatusCode,extracted)}
 snapshot:=db.Snapshot();var event domain.Event;for _,v:=range snapshot.Events{event=v}
 if event.Status!="pending"||event.StartSeconds!=12||len(event.EvidenceIDs)!=1||event.EvidenceIDs[0]!="seg-1"{t.Fatalf("proposal lost evidence or timing: %+v",event)}
 body:=strings.NewReader(`{"decision":"accept","review_note":"Contexto revisado"}`)
 req,_:=http.NewRequest(http.MethodPost,ts.URL+"/api/projects/p/events/"+event.ID+"/review",body);req.Header.Set("Content-Type","application/json")
 reviewed,err:=http.DefaultClient.Do(req);if err!=nil{t.Fatal(err)};defer reviewed.Body.Close()
 var result domain.Event;if err=json.NewDecoder(reviewed.Body).Decode(&result);err!=nil{t.Fatal(err)}
 if reviewed.StatusCode!=http.StatusOK||result.Status!="accepted"||result.ReviewedAt.IsZero()||result.ReviewNote!="Contexto revisado"{t.Fatalf("review was not audited: status=%d event=%+v",reviewed.StatusCode,result)}
 duplicate:=strings.NewReader(`{"decision":"reject"}`);req,_=http.NewRequest(http.MethodPost,ts.URL+"/api/projects/p/events/"+event.ID+"/review",duplicate);req.Header.Set("Content-Type","application/json")
 again,err:=http.DefaultClient.Do(req);if err!=nil{t.Fatal(err)};defer again.Body.Close();if again.StatusCode!=http.StatusConflict{t.Fatalf("second review should conflict, got %d",again.StatusCode)}
}
func TestHasEventSignal(t *testing.T){
 if !hasEventSignal("La directora anuncia el comienzo del proyecto."){t.Fatal("expected Spanish event signal")}
 if !hasEventSignal("The team announced the launch."){t.Fatal("expected English event signal")}
 if hasEventSignal("La conferencia trata sobre energía solar."){t.Fatal("unexpected event signal")}
}
