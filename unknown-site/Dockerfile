FROM python:latest

WORKDIR /app

#Dependencies
COPY unknown-site/requirements.txt .
RUN pip install -r requirements.txt

#Content
COPY unknown-site/manage.py .
COPY unknown-site/mysite ./mysite
COPY unknown-site/crochet ./crochet
COPY --chmod=755 unknown-site/run.sh .

#Migrate + run
CMD ["./run.sh"]