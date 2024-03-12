package httpx

import (
	"net/http"
	"net/http/httptest"
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

func TestBindAndValidate(t *testing.T) {
	// Define our test cases
	testCases := []struct {
		testName       string
		buildContext   func() echo.Context
		inputStruct    interface{}
		expectedOutput error
	}{
		{
			testName: "ValidQueryParams",
			buildContext: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/?bandwidth=2", nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				return c
			},
			inputStruct: &struct {
				Bandwidth uint64 `hx_place:"query" hx_must:"true" hx_query_name:"bandwidth" hx_default:"1" hx_range:"1-10"`
			}{},
			expectedOutput: nil,
		},
		{
			testName: "MissingRequiredQueryParam",
			buildContext: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/?lostparam=2", nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				return c
			},
			inputStruct: &struct {
				Bandwidth *uint64 `hx_place:"query" hx_must:"true" hx_query_name:"bandwidth" hx_default:"1" hx_range:"1-10"`
			}{},
			expectedOutput: errors.New("missing query paramm, name:bandwidth, path:.Bandwidth"),
		},
		// Add more test cases as needed.
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			c := tc.buildContext()

			// Act - call our function to test
			err := BindAndValidate(c, tc.inputStruct)

			if err != nil {
				assert.Equal(t, tc.expectedOutput.Error(), err.Error())
			} else {
				assert.Equal(t, err, tc.expectedOutput)
			}
			// Assert - check the output is as expected

		})
	}
}
