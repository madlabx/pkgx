package httpx

import (
<<<<<<< HEAD
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
=======
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
>>>>>>> 491ef3b (do clean)
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
	"github.com/stretchr/testify/assert"
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

<<<<<<< HEAD
=======
var _ Unmarshaler = (*Fid)(nil)

type Fid struct {
	Id      int64
	AbsPath string // 绝对地址，给外置存储没有扫描完硬盘使用的，和 Id 互斥
}

func (p Fid) MarshalJSON() ([]byte, error) {
	if p.Id <= 0 {
		if len(p.AbsPath) > 0 && p.AbsPath[0] != '/' {
			return nil, errors.New("abs is empyt or not abs addr")
		}
		return []byte("\"" + p.AbsPath + "\""), nil
	}
	return fmt.Appendf(nil, `"%d"`, p.Id), nil
}
func (p *Fid) UnmarshalJSONX(data []byte) error {
	data = bytes.TrimPrefix(data, []byte{'"'})
	data = bytes.TrimSuffix(data, []byte{'"'})
	if len(data) == 0 {
		return nil
	}
	if data[0] == '/' {
		p.AbsPath = string(data)
		return nil
	}

	var num json.Number
	err := json.Unmarshal(data, &num)
	if err != nil {
		return err
	}
	id, err := num.Int64()
	if err != nil {
		return err
	}
	if id < 0 {
		return fmt.Errorf("fid cannot be negative: %d", id)
	}
	p.Id = id

	fmt.Printf("After Unmarshl, %v\n", p)
	return nil
}

func (p *Fid) UnmarshalParam(param string) error {
	return p.UnmarshalJSONX([]byte(param))
}

func checkInterface(v reflect.Value) {
	if _, ok := v.Interface().(json.Unmarshaler); ok {
		fmt.Printf("it is Unmarshaler\n")
	}
}

func TestCheckInterface(t *testing.T) {
	checkInterface(reflect.ValueOf(&Fid{}))
}

>>>>>>> 491ef3b (do clean)
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
	mockRequestWithHeaders := func(method, uri string, body io.Reader, header http.Header) handleFunc {
		return func() echo.Context {
			e := echo.New()
			req := httptest.NewRequest(method, uri, body)
			req.Header = header
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			return c
		}
	}
<<<<<<< HEAD
	as := func(expect, output any) any {
		return assert.Equal(t, expect, output)
	}
=======
	as := func(expect, output any) bool {
		return assert.Equal(t, expect, output)
	}

>>>>>>> 491ef3b (do clean)
	testCases := []struct {
		testName      string
		buildContext  handleFunc
		structFunc    func(any) any
		expectedError error
	}{
		{
<<<<<<< HEAD
=======
			testName:     "SupportPtrArrayYJMarshalerBody",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Fids":["36310271995822508"]}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Fids []*Fid
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(int64(36310271995822508), parsed.(*inputStruct).Fids[0].Id)
			},
			expectedError: nil,
		},
		{
			testName:     "SupportArrayYJMarshalerBody",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Fids":["36310271995822508"]}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Fids []Fid
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(int64(36310271995822508), parsed.(*inputStruct).Fids[0].Id)
			},
			expectedError: nil,
		},
		{
			testName:     "SupportEmptyMarshalerXBody",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Fid *Fid
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return assert.Nil(t, parsed.(*inputStruct).Fid)
			},
			expectedError: nil,
		},
		{
			testName:     "SupportMarshalerXBody",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Id":"1"}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Id Fid
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(int64(1), parsed.(*inputStruct).Id.Id)
			},
			expectedError: nil,
		},
		{
			testName:     "SupportPtrMarshalerXBody",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Id":"1"}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Id *Fid
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(int64(1), parsed.(*inputStruct).Id.Id)
			},
			expectedError: nil,
		},
		{
			testName:     "SupportPtrPtrMarshalerXBody",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Id":"1"}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Id **Fid //unsupport
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(int64(1), (*parsed.(*inputStruct).Id).Id)
			},
			expectedError: errors.New("invalid type for field, type:ptr, value:\"1\", value type:string, path:Id"),
		},
		{
			testName:     "SupportYJMarshalerQueryParam",
			buildContext: mockRequest(http.MethodGet, "/?Id=12", nil),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Path struct {
						Id Fid `hx_place:"query"`
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(int64(12), parsed.(*inputStruct).Path.Id.Id)
			},
			expectedError: nil,
		},
		{
			testName:     "SupportYJMarshalerHeader",
			buildContext: mockRequestWithHeaders(http.MethodGet, "/?Qfd=12", nil, http.Header{"Id": {"123"}}),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Path struct {
						Id Fid `hx_place:"header"`
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(int64(123), parsed.(*inputStruct).Path.Id.Id)
			},
			expectedError: nil,
		},
		{
>>>>>>> 491ef3b (do clean)
			testName:     "ValidQueryParams",
			buildContext: mockRequest(http.MethodGet, "/?bandwidth=2", nil),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth uint64 `hx_place:"query" hx_must:"true" hx_name:"bandwidth" hx_default:"1" hx_range:"1-10"`
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

			expectedError: errors.New("missing query parameter Bandwidth"),
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

			expectedError: errors.New("missing parameter Bandwidth"),
		},
		{
			testName:     "BodyAnonymousMember",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Name":"test", "PageNum":1, "SortOrder":"asec"}`)),
			structFunc: func(parsed any) any {

				type FilterStruct struct {
					FilterField    string
					FilterOperator string `hx_range:"eq,gt,lt,in" hx_default:"eq"`
					FilterValue    string
				}
				type PageStruct struct {
					PageNum  int
					PageSize int
					TotalNum int64
				}
				type SortStruct struct {
					SortField string
					SortOrder string
				}
				type inputStruct struct {
					PageStruct
					SortStruct
					FilterStruct
					Name    string `hx_must:"true"`
					Recurse bool
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as("asec", parsed.(*inputStruct).SortOrder)
			},

			expectedError: nil,
		},
<<<<<<< HEAD

=======
		{
			testName:     "BodyAnonymousMember2",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Name":"test", "PageNum":1, "SortOrder":"asec"}`)),
			structFunc: func(parsed any) any {
				type PageStruct struct {
					PageSize int
				}
				type inputStruct struct {
					PageStruct
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(0, parsed.(*inputStruct).PageSize)
			},

			expectedError: nil,
		},
>>>>>>> 491ef3b (do clean)
		{
			testName:     "BodyDeepAnonymousMember",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Name":"test", "PageNum":1, "SortOrder":"asec", "PageNum":2}`)),
			structFunc: func(parsed any) any {

				type FilterStruct struct {
					FilterField    string
					FilterOperator string `hx_range:"eq,gt,lt,in" hx_default:"eq"`
					FilterValue    string
				}
				type PageStruct struct {
					PageNum  int
					PageSize int
					TotalNum int64
				}
				type SortStruct struct {
					SortField string
					SortOrder string
					PageStruct
				}
				type inputStruct struct {
					//PageStruct
					SortStruct
					FilterStruct
					Name    string `hx_must:"true"`
					Recurse bool
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(2, parsed.(*inputStruct).PageNum)
			},

			expectedError: nil,
		},
		{
<<<<<<< HEAD
=======
			testName:     "BodyNilSlice",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Name":"test", "PageNum":1, "SortOrder":"asec", "PageNum":2}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Source string   `json:"Source"`
					TaskNo []string `json:"TaskNo"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return assert.Nil(t, parsed.(*inputStruct).TaskNo)
			},

			expectedError: nil,
		},
		{
>>>>>>> 491ef3b (do clean)
			testName:     "BodyEmptyArray",
			buildContext: mockRequest(http.MethodGet, "/", strings.NewReader(`{"Name":"test", "PageNum":1, "SortOrder":"asec", "PageNum":2}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth []uint64 `hx_must:"true" hx_range:"0-10"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(uint64(0), parsed.(*inputStruct).Bandwidth)
			},

			expectedError: errors.New("missing parameter Bandwidth"),
		},
		{
			testName:     "BodyEmpty",
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

			expectedError: errors.New("missing parameter Bandwidth"),
		},
		{
			testName:     "InHeader",
			buildContext: mockRequestWithHeaders(http.MethodGet, "/", nil, http.Header{"X-Client-Id": {"172.1.2.1"}}),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					ClientIp string `hx_name:"X-Client-ID" hx_place:"header" hx_must:"true"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as("172.1.2.1", parsed.(*inputStruct).ClientIp)
			},

			expectedError: nil,
		},
		{
			testName:     "InHeaderMiss",
			buildContext: mockRequestWithHeaders(http.MethodGet, "/", nil, http.Header{"ClientId": {"172.1.2.1"}}),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					ClientId string `hx_place:"header" hx_must:"true"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as("172.1.2.1", parsed.(*inputStruct).ClientId)
			},

			expectedError: errors.New("missing header parameter ClientId"),
		},
		{
			testName:     "InHeaderName",
			buildContext: mockRequestWithHeaders(http.MethodGet, "/", nil, http.Header{"X-Real-Ip": {"172.1.2.1"}}),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					ClientId string `hx_place:"header" hx_name:"X-Real-Ip" hx_must:"true"`
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as("172.1.2.1", parsed.(*inputStruct).ClientId)
			},

			expectedError: nil,
		},
		{
			testName:     "PatchSetBodyValues",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2", strings.NewReader(`{"Bandwidth":1, "Quality":{"Level":1.0}}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth uint64 `hx_must:"true" hx_range:"0-10"`
					Quality   struct {
						Level float64 `hx_must:"true"`
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(1.0, parsed.(*inputStruct).Quality.Level)
			},

			expectedError: nil,
		},
		{
			testName:     "TestDeepPtrMember",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{"Bandwidths":[1,2,3,4], "Quality":{"Level":1.0}}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					//Bandwidth []uint64 `hx_must:"false" hx_range:"0-10"`
					Bandwidths []uint64 `hx_must:"false" hx_range:"0-10"`
					Quality    struct {
<<<<<<< HEAD
						Level *float64 `hx_must:"false"`
=======
						Level  *float64 `hx_must:"false"`
						EmpPtr *int
>>>>>>> 491ef3b (do clean)
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}
<<<<<<< HEAD
				//as(parsed.(*inputStruct).Bandwidths, []uint64{1, 2, 3, 4})
				return as(1.0, *(parsed.(*inputStruct).Quality.Level))
=======
				return as([]uint64{1, 2, 3, 4}, parsed.(*inputStruct).Bandwidths) &&
					as(1.0, *(parsed.(*inputStruct).Quality.Level)) &&
					assert.Nil(t, parsed.(*inputStruct).Quality.EmpPtr)

			},

			expectedError: nil,
		},
		{
			testName:     "TestPtrStructMember",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{"Bandwidths":[1,2,3,4], "Quality":{"Level":1.0}}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					//Bandwidth []uint64 `hx_must:"false" hx_range:"0-10"`
					//Bandwidths []uint64 `hx_must:"false" hx_range:"0-10"`
					Quality *struct {
						Level  float64 `hx_must:"false"`
						EmpPtr int
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}
				return as(1.0, parsed.(*inputStruct).Quality.Level) &&
					as(0, parsed.(*inputStruct).Quality.EmpPtr)

>>>>>>> 491ef3b (do clean)
			},

			expectedError: nil,
		},
		{
			testName:     "TestUintArray",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{"Bandwidths":[1,2,3,4], "Quality":{"Level":1.0}}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					//Bandwidth []uint64 `hx_must:"false" hx_range:"0-10"`
					Bandwidths []uint64 `hx_must:"false" hx_range:"0-10"`
					Quality    struct {
						Level float64 `hx_must:"false"`
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}
				as(parsed.(*inputStruct).Bandwidths, []uint64{1, 2, 3, 4})
				return as(1.0, parsed.(*inputStruct).Quality.Level)
			},

			expectedError: nil,
		}, {
			testName:     "TestStringArray",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{"Dirs":["test/wefwe", "123dir"]}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					//Bandwidth []uint64 `hx_must:"false" hx_range:"0-10"`
					Dirs    []string `hx_must:"false"`
					Quality struct {
						Level float64 `hx_must:"false"`
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}
				return as(parsed.(*inputStruct).Dirs, []string{"test/wefwe", "123dir"})
			},

			expectedError: nil,
		},
		{
			testName:     "TestLongInt",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{"Level":1712652096}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Level int `hx_must:"false"`
				}

				if parsed == nil {
					return &inputStruct{}
				}
				return as(1712652096, parsed.(*inputStruct).Level)
			},

			expectedError: nil,
<<<<<<< HEAD
=======
		}, {
			testName:     "TestPtrMember",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{"Level":1712652096}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Level *int `hx_must:"true"`
				}

				if parsed == nil {
					return &inputStruct{}
				}
				return as(1712652096, *(parsed.(*inputStruct).Level))
			},

			expectedError: nil,
		},
		{
			testName:     "TestEmptyPtrMember",
			buildContext: mockRequest(http.MethodPatch, "/we?Level2=2.1", strings.NewReader(`{"Level1":1712652096}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Level *int
				}

				if parsed == nil {
					return &inputStruct{}
				}
				return assert.Nil(t, parsed.(*inputStruct).Level)
			},

			expectedError: nil,
>>>>>>> 491ef3b (do clean)
		},
		{
			testName: "TestStructArray",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{"Dirs":[
<<<<<<< HEAD
    {"Name":"./jonathantest","CreateTime":1712652096},
    {"Name": "123dir"}
    ]
    }`)),
=======
		{"Name":"./jonathantest","CreateTime":1712652096},
		{"Name": "123dir"}
		]
		}`)),
>>>>>>> 491ef3b (do clean)
			structFunc: func(parsed any) any {
				type NewStruct struct {
					Name       string
					CreateTime int64
				}
				type inputStruct struct {
					//Bandwidth []uint64 `hx_must:"false" hx_range:"0-10"`
					Dirs    []NewStruct `hx_must:"false"`
					Quality struct {
						Level float64 `hx_must:"false"`
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}
				return as([]NewStruct{{Name: "./jonathantest", CreateTime: 1712652096}, {Name: "123dir"}}, parsed.(*inputStruct).Dirs)
			},

			expectedError: nil,
		},
		{
			testName:     "TestStringArray1",
			buildContext: mockRequest(http.MethodPost, "/we?Level=2.1", nil),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					Bandwidth uint64 `hx_must:"false" hx_range:"0-10"`
					Quality   struct {
						Level float64 `hx_must:"true"`
					}
				}

				if parsed == nil {
					return &inputStruct{}
				}

				return as(2.1, parsed.(*inputStruct).Quality.Level)
			},

			expectedError: nil,
		},
		{
			testName:     "TestArrayEmpty",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{"Paths":[], "Quality":{"Level":1.0}}`)),
			structFunc: func(parsed any) any {
				type inputStruct struct {
					//Bandwidth []uint64 `hx_must:"false" hx_range:"0-10"`
					Paths []string `hx_must:"true"`
				}

				if parsed == nil {
					return &inputStruct{}
				}
				//as(parsed.(*inputStruct).Bandwidths, []uint64{1, 2, 3, 4})
				return nil
			},

			expectedError: errors.New("missing parameter Paths"),
		},
		{
			testName:     "TestAny",
			buildContext: mockRequest(http.MethodPatch, "/we?Level=2.1", strings.NewReader(`{ "Quality":{"Level":1.0}}`)),
			structFunc: func(parsed any) any {
				type QualityType struct {
					Level float64
				}
				type inputStruct struct {
					Quality any
				}

				if parsed == nil {
					return &inputStruct{}
				}
				as(json.Number("1.0"), parsed.(*inputStruct).Quality.(map[string]interface{})["Level"])
				return nil
			},

			expectedError: nil,
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
				assert.Equal(t, tc.expectedError, err)
				tc.structFunc(input)
			}
			// Assert - check the output is as expected

		})
	}
}
<<<<<<< HEAD
=======

type testStruct struct {
	Name *string
	Age  *int
}

func TestTTT(t *testing.T) {
	target := &testStruct{}

	assert.Nil(t, json.Unmarshal([]byte(`{"Name": "123"}`), target))

	assert.Nil(t, target.Age)
}
>>>>>>> 491ef3b (do clean)
