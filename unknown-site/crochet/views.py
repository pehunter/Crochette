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

    # username, password = (body["username"], body["password"])

    # Issue login request
    print(body["username"], body["password"])
    loggedIn = requests.post("http://localhost:3000/login", json=body)
    print(loggedIn.text)

    if loggedIn.status_code == http.HTTPStatus.UNAUTHORIZED:
        return HttpResponse(
            "This user is already logged in.", status=http.HTTPStatus.UNAUTHORIZED
        )
    if loggedIn.status_code != http.HTTPStatus.OK:
        return HttpResponse(
            loggedIn.status_code,
            "An error occurred trying to log in.",
            status=loggedIn.status_code,
        )
    else:
        return HttpResponse(loggedIn.text)


def addUser(request: HttpRequest):
    # Create the user
    newUser = User()

    # Save them
    newUser.save()

    return HttpResponse(f"New user created @{newUser.id}")


def getUsers(request: HttpRequest):
    users = list(User.objects.all())
    json = {}
    for user in users:
        json[user.pk] = user.name
    return JsonResponse(json)
