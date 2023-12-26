package ginadapter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GinLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.New(r)

			req := events.APIGatewayProxyRequest{
				Path:       "/ping",
				HTTPMethod: "GET",
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))

			resp, err = adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})

var _ = Describe("GinLambdaV2 tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.NewV2(r)

			req := events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: "GET",
						Path:   "/ping",
					},
				},
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))

			resp, err = adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})

var _ = Describe("GinLambdaALB tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.NewALB(r)

			req := events.ALBTargetGroupRequest{
				HTTPMethod: "GET",
				Path:       "/ping",
				RequestContext: events.ALBTargetGroupRequestContext{
					ELB: events.ELBContext{TargetGroupArn: " ad"},
				}}

			resp, err := adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))

			resp, err = adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Context("Decoding query parameters", func() {
		type tQuery struct {
			Email string `url:"email" form:"email" json:"email"`
		}
		type tResp struct {
			Error    string `json:"error"`
			Received string `json:"received"`
			Decoded  string `json:"decoded"`
		}

		It("Decodes query parameters from request correctly", func() {
			log.Println("Starting test")

			testCase := "some@site.com"
			r := gin.Default()
			r.GET("/users", func(c *gin.Context) {
				log.Println("In the handler!!")

				var q tQuery
				err := c.ShouldBindQuery(&q)
				if err != nil {
					log.Printf("failed bind query: %s", err)
					c.JSON(200, tResp{
						Error:    fmt.Sprintf("failed bind query: %s", err),
						Received: q.Email,
					})
					return
				}

				decoded, err := url.QueryUnescape(q.Email)
				if err != nil {
					log.Printf("failed to decode query parameter: %s", err)
					c.JSON(200, tResp{
						Error:    fmt.Sprintf("failed to decode query parameter: %s", err),
						Received: q.Email,
						Decoded:  decoded,
					})
					return
				}

				if q.Email != testCase {
					log.Printf("parameter '%s is not as expected '%s'", q.Email, testCase)
					c.JSON(200, tResp{
						Error:    "parameter is not as expected",
						Received: q.Email,
						Decoded:  decoded,
					})
					return
				}

				c.JSON(200, tResp{
					Received: q.Email,
					Decoded:  decoded,
				})
			})

			adapter := ginadapter.NewALB(r)

			req := events.ALBTargetGroupRequest{
				HTTPMethod: "GET",
				Path:       "/users",
				QueryStringParameters: map[string]string{
					"email": url.QueryEscape(testCase),
				},
				RequestContext: events.ALBTargetGroupRequestContext{
					ELB: events.ELBContext{TargetGroupArn: " ad"},
				}}

			resp, err := adapter.Proxy(req)

			log.Printf("Body: %s\n", resp.Body)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Body).To(Not(BeEmpty()))

			var result tResp
			err = json.Unmarshal([]byte(resp.Body), &result)
			Expect(err).To(BeNil())

			Expect(result.Received).To(Equal(testCase), fmt.Sprintf("expected: '%s', got: '%s', it should decoded: '%s'", testCase, result.Received, result.Decoded))
			Expect(result.Error).To(BeEmpty(), result.Error)
		})
	})
})
