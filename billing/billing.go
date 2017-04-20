/*
Blah blah.

Handler

Blah blah.

Manual Authentication

Blah blah.
*/
package billing

import (
  "net/http"
  "time"
  "strings"
  "encoding/json"
  "bytes"
  "io/ioutil"
  "fmt"
)

type cookie struct {
  token string `json:"token"`
  logoutUrl string `json:"logout_url"`
}
type space struct {
  LogoutUrl string
}

type plasso struct {
  LoggedIn bool
  Token string
  Id string
  PlanId int32
  Space space
}

type gqlQuery struct {
  Query string `json:"query"`
  Variables map[string]string `json:"variables"`
}

type gqlResponse struct {
  Data struct {
    Member struct {
      Id string `json:"id"`
      PlanId int32 `json:"planId"`
      Space struct {
        Slug string `json:"slug"`
        LogoutUrl string `json:"logoutUrl"`
      } `json:"space"`
    } `json:"member"`
  } `json:"data"`
}

type handler func(http.ResponseWriter, *http.Request)

func New(token string) (*plasso, error) {
  var client = &http.Client{
    Timeout: 1 * time.Second,
  }

  var template = "{member(token: $token){id,planId,space{logoutUrl}}}"
  var gql = gqlQuery{query, {"token": token}}

  body, err := json.Marshal(gql)
  if err != nil {
    return nil, err
  }

  req, err := http.NewRequest("POST", "https://api.plasso.com", bytes.NewBuffer(body))
  if err != nil {
    return nil, err
  }
  req.Header.Set("Content-Type", "application/json")

  res, err := client.Do(req)
  if err != nil {
    return nil, err
  }
  defer res.Body.Close()

  responseBody, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, err
  }

  var r gqlResponse
  err = json.Unmarshal(responseBody, &r)
  if err != nil {
    return nil, err
  }

  var m = r.Data.Member
  return &plasso{true, token, m.Id, m.PlanId, space{ m.Space.LogoutUrl }}, nil
}

func fromRequest(r *http.Request) (*plasso, error) {
  // If cookie exists
    // Parse it into plasso object
  // If cookie does not exists
    // Look for token get param
    // if logout return nil
}

func (p *plasso) Protect(handler handler) handler {
  return func (w http.ResponseWriter, r *http.Request) {
    plasso, err := FromRequest(r)
    if err != nil {
      // Redirect to root of host
    }
    if plasso.LoggedOut {
      logout(w);// Redirect to logoutUrl
      return;
    }
    
  }
}

func GetData(token string) error {

}

func UpdateSettings(token string) error {

}

func UpdateCreditCard(token string) error {

}

func Delete(token string) error {

}

func Logout(token string) error {

}

