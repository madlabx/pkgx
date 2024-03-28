package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
	"github.com/stretchr/testify/mock"
)

// Here we'll define a mock echo.Context
type MockContext struct {
	echo.Context
	mock.Mock
}

func (m *MockContext) QueryParam(name string) string {
	args := m.Called(name)
	return args.String(0)
}

func (m *MockContext) Bind(i interface{}) error {
	args := m.Called(i)
	return args.Error(0)
}

type handleFunc func() echo.Context

func TestBindAndValidate(t *testing.T) {
	// Define our test cases
	mockRequest := func(method, uri string, body io.Reader) handleFunc {
		return func() echo.Context {
			e := echo.New()
			req := httptest.NewRequest(method, uri, body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			return c
		}
	}
	as := func(expect, output any) bool {
		return assert.Equal(t, expect, output)
	}
	testCases := []struct {
		testName      string
		buildContext  handleFunc
		structFunc    func(any) any
		expectedError error
	}{
		{
			testName:     "ValidQueryParams",
			buildContext: mockRequest(http.MethodGet, "/?bandwidth=2", nil),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth uint64 `hx_place:"query" hx_must:"true" hx_query_name:"bandwidth" hx_default:"1" hx_range:"1-10"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(uint64(2), parsed.(*inputStruct).Bandwidth)
			},
			expectedError: nil,
		},
		{
			testName:     "MissingRequiredQueryParam",
			buildContext: mockRequest(http.MethodGet, "/?lostparam=2", nil),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth uint64 `hx_place:"query" hx_must:"true" hx_query_name:"bandwidth" hx_default:"1" hx_range:"1-10"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(uint64(2), parsed.(*inputStruct).Bandwidth)
			},

			expectedError: errors.New("missing param Bandwidth"),
		},

		{
			testName:     "BodyInt64Value",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Bandwidth":1}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth uint64 `hx_must:"true" hx_default:"1" hx_range:"1-10"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(uint64(1), parsed.(*inputStruct).Bandwidth)
			},

			expectedError: nil,
		},

		{
			testName:     "BodyIntLostMust",
			buildContext: mockRequest(http.MethodGet, "/", nil),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth uint64 `hx_must:"true" hx_range:"0-10"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(uint64(0), parsed.(*inputStruct).Bandwidth)
			},

			expectedError: errors.New("missing param Bandwidth"),
		},

		{
			testName:     "MissingMustParamInDeepStruct",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Bandwidth":1}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth uint64 `hx_must:"true" hx_range:"0-10"`
					Quality   struct {
						Level int `hx_must:"true"`
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(uint64(0), parsed.(*inputStruct).Bandwidth)
			},

			expectedError: errors.New("missing param Quality.Level"),
		},
		// Add more test cases as needed.
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			c := tc.buildContext()

			// Act - call our function to test
			input := tc.structFunc(nil)
			err := BindAndValidate(c, input)

			if err != nil {
				if tc.expectedError == nil {
					assert.Nil(t, err)
					return
				}
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				tc.structFunc(input)
			}
			// Assert - check the output is as expected

		})
	}
}
