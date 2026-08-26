FROM python:latest

WORKDIR /app

#Prevents Python from writing pyc files to disk
ENV PYTHONDONTWRITEBYTECODE=1
#Prevents Python from buffering stdout and stderr
ENV PYTHONUNBUFFERED=1 
 
#Dependencies
RUN pip install --upgrade pip
COPY unknown-site/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

#Content
COPY unknown-site/manage.py .
COPY unknown-site/mysite ./mysite
COPY unknown-site/crochet ./crochet
COPY --chmod=755 unknown-site/run.sh .

#Migrate + run
EXPOSE 8000
CMD ["./run.sh"]