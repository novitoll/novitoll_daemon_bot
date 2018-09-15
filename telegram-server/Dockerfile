FROM centos

RUN yum update -y && \
  yum install -y python-pip

COPY requirements.txt /opt/vahter/

RUN pip install -r /opt/vahter/requirements.txt

VOLUME /opt/vahter
WORKDIR /opt/vahter

EXPOSE 5001
ENTRYPOINT ["python", "app.py", "${TELEGRAM_BOT_TOKEN}"]