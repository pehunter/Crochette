from django.db import models
from django.db.models import CASCADE


# Create your models here.
class User(models.Model):
    # Name
    name = models.CharField(max_length=40)

    # Password
    password = models.CharField(max_length=40)

    def __str__(self):
        return self.name


class Pattern(models.Model):
    # Name
    name = models.CharField(max_length=100)

    # Creator; who created this pattern
    creator = models.ForeignKey(User, on_delete=CASCADE)

    # Steps; an encoding of what steps to follow
    steps = models.CharField(max_length=1000)

    def __str__(self):
        return self.name


class Progress(models.Model):
    # The pattern represented by this progress
    pattern = models.ForeignKey(Pattern, on_delete=CASCADE)

    # The user that would be following this pattern
    user = models.ForeignKey(User, on_delete=CASCADE)

    # Amount of progress made
    progress = models.IntegerField(default=0)

    def __str__(self):
        return f"{self.pattern.name} ({self.progress})"
