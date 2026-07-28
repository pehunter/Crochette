from django.db import models


# Create your models here.
class Light(models.Model):
    # Name
    name = models.CharField(max_length=40)
    # 0% to 100%
    brightness = models.FloatField(default=1.0)

    # 0 to 255
    r = models.SmallIntegerField(default=255)
    g = models.SmallIntegerField(default=255)
    b = models.SmallIntegerField(default=255)

    # Might add other ways to identify this light but worry about later
    def getColorAsCSS(self) -> str:
        return f"#{self.r:x}{self.g:x}{self.b:x}"

    def getBrightnessAsPercent(self) -> int:
        return int(self.brightness * 100)

    def __str__(self) -> str:
        return self.name

    # Parse a name string, or return None if the conversion fails
    @staticmethod
    def parseName(name: str):
        if len(name) > 40:
            return None
        return name

    # Parse a color string into three integer values, or return None if the conversion fails
    @staticmethod
    def parseColor(color: str) -> tuple[int, int, int] | None:
        splitGroups = [int(color[i : i + 2], 16) for i in range(1, len(color), 2)]
        for group in splitGroups:
            if group < 0 or group > 255:
                return None
        if len(splitGroups) == 3:
            return (splitGroups[0], splitGroups[1], splitGroups[2])
        else:
            return None

    # Parse a brightness string into a float, or return None if the conversion fails
    @staticmethod
    def parseBrightness(brightness: str) -> float | None:
        brightnessVal = int(brightness) / 100.0
        if 1 >= brightnessVal >= 0:
            return brightnessVal
        return None
