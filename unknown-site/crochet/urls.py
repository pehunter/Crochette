from django.urls import path
from . import views

app_name = "crochet"
urlpatterns = [
    path("cookie/", views.cookie, name="cookie"),
    path("login/", views.login, name="login"),
    path("register/", views.register, name="register"),
    path("user/<int:id>/patterns/", views.userPatterns, name="userPatterns"),
    path("user/<int:id>/progress/", views.userProgress, name="userProgress"),
    path("pattern/<int:id>", views.patternDetail, name="patternDetail"),
    path("pattern/", views.createPattern, name="createPattern"),
    path("progress/<int:id>", views.progressDetail, name="progressDetail"),
    path("progress/", views.manageProgress, name="manageProgress"),
]
