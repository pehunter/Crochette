from django.urls import path
from . import views

app_name = "crochet"
urlpatterns = [
    path("adduser/", views.addUser, name="adduser"),
    path("users/", views.getUsers, name="users"),
]
