package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var WebhookUrl string

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Embed struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Url         string  `json:"url,omitempty"`
	Fields      []Field `json:"fields,omitempty"`
}

type Message struct {
	Embeds []Embed `json:"embeds"`
}

type Gql struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
}

func init() {
	flag.StringVar(&WebhookUrl, "url", "", "Webhook URL")
	flag.Parse()
}

func main() {
	fields := fetch()
	send(fields)
}

func fetch() []Field {
	msg := &Gql{
		Query: `
	query promotionsQuery($namespace: String!, $country: String!, $locale: String!) {
		Catalog {
			catalogOffers(
				namespace: $namespace
				locale: $locale
				params: {
					category: "freegames"
					country: $country
					sortBy: "effectiveDate"
					sortDir: "asc"
				}
			) {
				elements {
					title
					description
					id
					namespace
					categories {
						path
					}
					linkedOfferNs
					linkedOfferId
					keyImages {
						type
						url
					}
					productSlug
					promotions {
						promotionalOffers {
							promotionalOffers {
								startDate
								endDate
								discountSetting {
									discountType
									discountPercentage
								}
							}
						}
						upcomingPromotionalOffers {
							promotionalOffers {
								startDate
								endDate
								discountSetting {
									discountType
									discountPercentage
								}
							}
						}
					}
				}
			}
		}
	}        
`,
		Variables: map[string]string{
			"country": "GB",
			"locale": "en-US",
			"namespace": "epic",
		},
	}

	body := post("https://graphql.epicgames.com/graphql", msg)
	var result map[string]map[string]map[string]map[string]interface{}
	json.Unmarshal(body, &result)

	url := "https://www.epicgames.com/store/en-US/product/"
	items := result["data"]["Catalog"]["catalogOffers"]["elements"].([]interface{})
	var fields []Field

	for _, item := range items {
		promos := item.(map[string]interface{})["promotions"]
		offers := promos.(map[string]interface{})["promotionalOffers"]

		if len(offers.([]interface{})) == 0 {
			continue
		}

		name := fmt.Sprint(item.(map[string]interface{})["title"])
		slug :=  strings.Join([]string{
			url,
			fmt.Sprint(item.(map[string]interface{})["productSlug"]),
		}, "")

		fields = append(fields, Field{
			Name:  name,
			Value: slug,
		})
	}

	return fields
}

func send(fields []Field) {
	msg := &Message{
		Embeds: []Embed{
			{
				Title:       "Free on Epic",
				Description: "Todays free games are...",
				Url:         "https://www.epicgames.com/store/en-US/free-games",
				Fields:      fields,
			},
		},
	}

	body := post(WebhookUrl, msg)
	fmt.Println(string(body))
}

func post(url string, msg interface{}) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body
}
