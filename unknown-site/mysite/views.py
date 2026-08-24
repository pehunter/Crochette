import http

from django.http import HttpRequest, HttpResponse


def health(request: HttpRequest):
    return HttpResponse(status=http.HTTPStatus.OK)
