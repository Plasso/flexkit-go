/*
This package allows you to interact with your flexkit space.  This includes
authenticating a customer, seeing plan details, your data items associated with the
customer, updating payment details, subscribing them to plans, and purchasing products.

Example

For example to authenticate:
	package main

	import (
		fk "github.com/Plasso/plasso-go/flexkit"
	)

	func main() {
		var member, err = fk.Login(fk.LoginRequest{PublicKey: "test", Email: "mike+1@plasso.com", Password: "password"})
		if err != nil {
			fmt.Println(err)
			return
		}
		memberData, err := member.GetData()
		if err != nil {
			fmt.Println(err)
			return
		}
		// memberData.Plans
		// memberData.Name
		// memberData.Id
		// memberData....
	}

*/
package flexkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const domain string = "https://plasso.com"

const getMemberQuery string = `
query getMember($token: String) {
  member(token: $token) {
  	id,
    name,
    email,
    ccType,
    ccLast4,
    shippingInfo {
      name
      address
      city
      state
      zip
      country
    },
    dataFields {
      id,
      value
    },
    plan {
    	alias
    }
  }
}`

type gqlQuery struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
}

type memberDataResponse struct {
	Data struct {
		Member struct {
			Id      string `json:"id"`
			Name    string `json:"name"`
			Email   string `json:"email"`
			CcType  string `json:"ccType"`
			CcLast4 string `json:"ccLast4"`
			Plan    struct {
				Alias string `json:"alias"`
			} `json:"plan"`
			ShippingInfo struct {
				Name    string `json:"name"`
				Address string `json:"address"`
				City    string `json:"city"`
				State   string `json:"state"`
				Zip     string `json:"zip"`
				Country string `json:"country"`
			} `json:"shippingInfo"`
			DataFields []DataItem `json:"dataFields"`
		} `json:"member"`
	} `json:"data"`
}

// The structure that should be filled out and passed to the Login function.
type LoginRequest struct {
	PublicKey string `json:"public_key"` // Public Key of Plasso user
	Email     string `json:"email"`      // Email customer uses to log in
	Password  string `json:"password"`   // Password of customer
}

// This structure represents a product.
type Product struct {
	Id     string `json:"id"`     // Plasso product id
	Qty    string `json:"qty"`    // Quantity
	Amount string `json:"amount"` // Amount for variable price products
}

// The structure that should be filled out and passed to the CreatePayment function.
type PaymentRequest struct {
	PublicKey       string     `json:"public_key"`       // Plasso customer public key
	Token           string     `json:"token"`            // Token returned from javascript flexkit GetToken call
	Products        []Product  `json:"products"`         // List of products
	BillingAddress  string     `json:"billing_address"`  // Billing address of customer (optional depending on plan).
	BillingCity     string     `json:"billing_city"`     // Billing city of customer (optional depending on plan).
	BillingState    string     `json:"billing_state"`    // Billing state of customer (optional depending on plan).
	BillingZip      string     `json:"billing_zip"`      // Billing zip of customer (optional depending on plan).
	BillingCountry  string     `json:"billing_country"`  // Billing country of customer (optional depending on plan).
	ShippingName    string     `json:"shipping_name"`    // Shipping name of customer (optional depending on plan).
	ShippingAddress string     `json:"shipping_address"` // Shipping address of customer (optional depending on plan).
	ShippingCity    string     `json:"shipping_city"`    // Shipping city of customer (optional depending on plan).
	ShippingState   string     `json:"shipping_state"`   // Shipping state of customer (optional depending on plan).
	ShippingZip     string     `json:"shipping_zip"`     // Shipping zip of customer (optional depending on plan).
	ShippingCountry string     `json:"shipping_country"` // Shipping country of customer (optional depending on plan).
	ShippingOptions string     `json:"shipping_options"` // Shipping options of customer (optional depending on plan).
	DataFields      []DataItem `json:"data_fields"`      // Data items (optional)
	Coupon          string     `json:"coupon"`           // Coupon code (optional)
	Email           string     `json:"email"`            // Email customer provided
	Name            string     `json:"name"`             // Name of customer
}

// Represents a data item
type DataItem struct {
	Id    string `json:"id"`    // The id of the data item
	Value string `json:"value"` // The value of the data item
}

// The structure that should be filled out and passed to the CreateSubscription function.
type SubscriptionRequest struct {
	SubscriptionFor string     `json:"subscription_for"`
	Email           string     `json:"email"`            // Email customer provided
	Name            string     `json:"name"`             // Name of customer
	Password        string     `json:"password"`         // Customer Password
	Plan            string     `json:"plan"`             // The plan id you are subscribing to
	Token           string     `json:"token"`            // Token returned from javascript flexkit GetToken call
	BillingAddress  string     `json:"billing_address"`  // Billing address of customer (optional depending on plan).
	BillingCity     string     `json:"billing_city"`     // Billing city of customer (optional depending on plan).
	BillingState    string     `json:"billing_state"`    // Billing state of customer (optional depending on plan).
	BillingZip      string     `json:"billing_zip"`      // Billing zip of customer (optional depending on plan).
	BillingCountry  string     `json:"billing_country"`  // Billing country of customer (optional depending on plan).
	ShippingName    string     `json:"shipping_name"`    // Shipping name of customer (optional depending on plan).
	ShippingAddress string     `json:"shipping_address"` // Shipping address of customer (optional depending on plan).
	ShippingCity    string     `json:"shipping_city"`    // Shipping city of customer (optional depending on plan).
	ShippingState   string     `json:"shipping_state"`   // Shipping state of customer (optional depending on plan).
	ShippingZip     string     `json:"shipping_zip"`     // Shipping zip of customer (optional depending on plan).
	ShippingCountry string     `json:"shipping_country"` // Shipping country of customer (optional depending on plan).
	ShippingOptions string     `json:"shipping_options"` // Shipping options of customer (optional depending on plan).
	DataFields      []DataItem `json:"data_fields"`      // Data items (optional)
	PublicKey       string     `json:"public_key"`       // Plasso customer public key
}

type tokenResponse struct {
	Token string `json:"token"`
}

// A request to update a members payment information
type CreditCardRequest struct {
	Last4       string `json:"cc_last_4"` // Informational, Last 4 of credit card
	Type        string `json:"cc_type"`   // Informational, type of card
	PlanId      string `json:"plan"`      // Allows changing plan
	Token       string `json:"token"`     // Stripe source token
	memberToken string `json:"pltoken"`
}

// A request to change a members settings
type SettingsRequest struct {
	Email           string `json:"email"`            // Email customer provided
	Name            string `json:"name"`             // Name of customer
	ShippingName    string `json:"shipping_name"`    // Shipping name of customer (optional depending on plan).
	ShippingAddress string `json:"shipping_address"` // Shipping address of customer (optional depending on plan).
	ShippingCity    string `json:"shipping_city"`    // Shipping city of customer (optional depending on plan).
	ShippingState   string `json:"shipping_state"`   // Shipping state of customer (optional depending on plan).
	ShippingZip     string `json:"shipping_zip"`     // Shipping zip of customer (optional depending on plan).
	ShippingCountry string `json:"shipping_country"` // Shipping country of customer (optional depending on plan).
	ShippingOptions string `json:"shipping_options"` // Shipping options of customer (optional depending on plan).
	token           string `json:"pltoken"`
}

// A handle to a member
type Member struct {
	PublicKey string // Public key of Plasso user
	Token     string // This token changes after every login
}

// Information about a member
type MemberData struct {
	Id              string     // A unique id identifying the user, does not change
	Email           string     // Email customer provided
	Name            string     // Name of customer
	CreditCardLast4 string     // Informational, Last 4 of credit card
	CreditCardType  string     // Informational, type of card
	ShippingName    string     // Shipping name of customer (optional depending on plan).
	ShippingAddress string     // Shipping address of customer (optional depending on plan).
	ShippingCity    string     // Shipping city of customer (optional depending on plan).
	ShippingState   string     // Shipping state of customer (optional depending on plan).
	ShippingZip     string     // Shipping zip of customer (optional depending on plan).
	ShippingCountry string     // Shipping country of customer (optional depending on plan).
	ShippingOptions string     // Shipping options of customer (optional depending on plan).
	DataFields      []DataItem // Data items (optional)
	Plan            string     // Plan ID
}

func graphQL(query string, variables map[string]string, response interface{}) error {
	var client = &http.Client{
		Timeout: 15 * time.Second,
	}

	var gql = gqlQuery{query, variables}

	body, err := json.Marshal(gql)
	if err != nil {
		return err
	}

	var url = fmt.Sprintf("%s/graphql", domain)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(responseBody, response)
}

func sendRequest(kind string, path string, request interface{}) ([]byte, error) {
	var url = fmt.Sprintf("%s%s", domain, path)
	var client = &http.Client{
		Timeout: 30 * time.Second,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(kind, url, bytes.NewBuffer(body))
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

	if res.StatusCode < 200 || res.StatusCode > 299 {
		var errorText = fmt.Sprintf(
			"%s %d %s %s",
			kind,
			res.StatusCode,
			url,
			string(responseBody))
		return responseBody, errors.New(errorText)
	}

	return responseBody, nil
}

// Authenticates and returns a Member.
func Login(request LoginRequest) (*Member, error) {
	body, err := sendRequest("POST", "/api/service/login", request)
	if err != nil {
		return nil, err
	}

	var r tokenResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	return &Member{request.PublicKey, r.Token}, nil
}

// Get member details
func (member *Member) GetData() (*MemberData, error) {
	var response memberDataResponse
	var variables = map[string]string{"token": member.Token}
	var memberData MemberData

	var err = graphQL(getMemberQuery, variables, &response)
	if err != nil {
		return nil, err
	}

	memberData.CreditCardLast4 = response.Data.Member.CcLast4
	memberData.CreditCardType = response.Data.Member.CcType
	memberData.DataFields = response.Data.Member.DataFields
	memberData.Email = response.Data.Member.Email
	memberData.Id = response.Data.Member.Id
	memberData.Name = response.Data.Member.Name
	memberData.Plan = response.Data.Member.Plan.Alias
	memberData.ShippingAddress = response.Data.Member.ShippingInfo.Address
	memberData.ShippingCity = response.Data.Member.ShippingInfo.City
	memberData.ShippingCountry = response.Data.Member.ShippingInfo.Country
	memberData.ShippingName = response.Data.Member.ShippingInfo.Name
	memberData.ShippingState = response.Data.Member.ShippingInfo.State
	memberData.ShippingZip = response.Data.Member.ShippingInfo.Zip

	return &memberData, nil
}

// Update member settings
func (member *Member) UpdateSettings(request SettingsRequest) error {
	request.token = member.Token
	_, err := sendRequest("POST", "/api/services/user?action=settings", request)
	if err != nil {
		return err
	}

	return nil
}

// Update members payment details
func (member *Member) UpdateCreditCard(request CreditCardRequest) error {
	request.memberToken = member.Token
	_, err := sendRequest("POST", "/api/services/user?action=cc", request)
	if err != nil {
		return err
	}

	return nil
}

// Creates a new payment
func CreatePayment(request PaymentRequest) error {
	_, err := sendRequest("POST", "/api/payments", request)
	if err != nil {
		return err
	}

	return nil
}

// Creates a new subscription to a plan
func CreateSubscription(request SubscriptionRequest) (*Member, error) {
	request.SubscriptionFor = "space"
	body, err := sendRequest("POST", "/api/subscriptions", request)
	if err != nil {
		return nil, err
	}

	var r tokenResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	return &Member{request.PublicKey, r.Token}, nil
}

// Deletes the member.  The member object cannot be used after this call and must be recreated.
func (member *Member) Delete() error {
	var request = map[string]string{"token": member.Token}

	_, err := sendRequest("DELETE", "/api/service/user", request)
	if err != nil {
		return err
	}

	return nil
}

// Logs out the member.  The member object cannot be used after this call and must be recreated.
func (member *Member) Logout() error {
	var request = map[string]string{"token": member.Token, "public_key": member.PublicKey}

	_, err := sendRequest("POST", "/api/service/logout", request)
	if err != nil {
		return err
	}

	return err
}
