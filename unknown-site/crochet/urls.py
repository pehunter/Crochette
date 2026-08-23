from django.urls import path
from . import views

app_name = "crochet"
urlpatterns = [
    path("cookie/", views.cookie, name="cookie"),
    path("login/", views.login, name="login"),
    path("users/", views.getUsers, name="users"),
]
