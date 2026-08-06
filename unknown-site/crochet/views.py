from django.http import HttpRequest, HttpResponse, JsonResponse
from django.shortcuts import render

# Create your views here.
from .models import Pattern, User


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
