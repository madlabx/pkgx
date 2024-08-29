package httpx

type JsonResponseWrapper interface {
	ToHttpXJsonResponse() *JsonResponse
}
