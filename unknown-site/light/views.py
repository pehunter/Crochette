from django.http import HttpRequest, HttpResponse, HttpResponseRedirect
from django.shortcuts import render
from django.template import loader
from django.urls import reverse

from light.models import Light


# Create your views here.
def index(request: HttpRequest):
    lights = Light.objects.all()
    return render(request, "light/index.html", {"lights": lights})


def update(request: HttpRequest, id: int):
    try:
        print(request.POST)
        # Pull
        name = request.POST["name"]
        rgb = request.POST["color"]
        brightness = request.POST["brightness"]

        # Verify
        if not isinstance(name, str):
            raise ValueError()
        name = Light.parseName(name)

        if not isinstance(rgb, str) or not Light.parseColor(rgb):
            raise ValueError()
        rgb = Light.parseColor(rgb)

        if not isinstance(brightness, str) or not Light.parseBrightness(brightness):
            raise ValueError()
        brightness = Light.parseBrightness(brightness)

        print(name, rgb, brightness)
        if not name or not rgb or not brightness:
            raise ValueError()

        # Retrieve light
        selected_light = Light.objects.get(pk=id)
    except (KeyError, ValueError, Light.DoesNotExist):
        return HttpResponse("Could not update the light.")
    else:
        selected_light.name = name
        selected_light.r = rgb[0]
        selected_light.g = rgb[1]
        selected_light.b = rgb[2]
        selected_light.brightness = brightness

        # save it
        selected_light.save()
        return HttpResponseRedirect(reverse("light:index"))
