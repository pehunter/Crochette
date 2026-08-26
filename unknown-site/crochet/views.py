import http
import json

import django.middleware.csrf
from django.http import HttpRequest, HttpResponse, JsonResponse
from django.shortcuts import render
from django.views.decorators.http import require_POST
import requests

# Create your views here.
from .models import Pattern, User


def cookie(request: HttpRequest):
    return HttpResponse(
        "u want that cookie? %s" % django.middleware.csrf.get_token(request=request)
    )


@require_POST
def login(request: HttpRequest):
    body = json.loads(request.body)

    # Verify body contains correct parameters
    if "username" not in body or "password" not in body:
        return HttpResponse(
            http.HTTPStatus.BAD_REQUEST, "Request was not formatted properly"
        )

    # Issue login request
    loggedIn = requests.post("http://sessions:8080/login", json=body)

    if loggedIn.status_code == http.HTTPStatus.UNAUTHORIZED:
        return HttpResponse(
            "This user is already logged in.", status=http.HTTPStatus.UNAUTHORIZED
        )
    if loggedIn.status_code != http.HTTPStatus.OK:
        return HttpResponse(
            "An error occurred trying to log in.",
            status=loggedIn.status_code,
        )
    else:
        return HttpResponse(loggedIn.text)


@require_POST
def register(request: HttpRequest):
    body = json.loads(request.body)

    # Verify body contains correct parameters
    if "username" not in body or "password" not in body:
        return HttpResponse(
            http.HTTPStatus.BAD_REQUEST, "Request was not formatted properly"
        )

    # Save them
    response = requests.post("http://userinfo:8080/register", json=body)
    respBody = response.json()

    match response.status_code:
        case http.HTTPStatus.CONFLICT:
            return HttpResponse(
                "A user with this name already exists.", status=http.HTTPStatus.CONFLICT
            )
        case http.HTTPStatus.OK:
            return HttpResponse(respBody["id"])
        case _:
            print(respBody["error"])
            return HttpResponse("An internal server error occurred.", status=500)


def checkLoggedIn(body: dict) -> int:
    if "sessionkey" not in body:
        return -1

    print(f"http://sessions:8080/${body["sessionkey"]}")
    isLoggedIn = requests.get(f"http://sessions:8080/{body["sessionkey"]}")

    if isLoggedIn.status_code != http.HTTPStatus.OK:
        print(isLoggedIn.text)
        return -1
    else:
        return int(isLoggedIn.text)


def userPatterns(request: HttpRequest, id: int):
    try:
        body = json.loads(request.body)
    except:
        return HttpResponse("No body was attached.", status=http.HTTPStatus.BAD_REQUEST)

    login = checkLoggedIn(body)
    if login == -1:
        return HttpResponse(
            "You are not currently logged in.", status=http.HTTPStatus.UNAUTHORIZED
        )

    response = requests.get(f"http://userinfo:8080/patterns/{id}")

    match response.status_code:
        case http.HTTPStatus.OK:
            return HttpResponse(response.text)
        case _:
            print(response.text)
            return HttpResponse(
                "An error occurred trying to retrieve patterns.",
                status=http.HTTPStatus.INTERNAL_SERVER_ERROR,
            )


def userProgress(request: HttpRequest, id: int):
    try:
        body = json.loads(request.body)
    except:
        return HttpResponse("No body was attached.", status=http.HTTPStatus.BAD_REQUEST)

    login = checkLoggedIn(body)
    if login == -1:
        return HttpResponse(
            "You are not currently logged in.", status=http.HTTPStatus.UNAUTHORIZED
        )

    response = requests.get(f"http://userinfo:8080/progress/{id}")

    match response.status_code:
        case http.HTTPStatus.OK:
            return HttpResponse(response.text)
        case _:
            print(response.text)
            return HttpResponse(
                "An error occurred trying to retrieve progress.",
                status=http.HTTPStatus.INTERNAL_SERVER_ERROR,
            )
