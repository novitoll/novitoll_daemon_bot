FROM centos

ENV PYTHONUNBUFFERED=0
ENV FLASK_APP=/opt/vahter/bot/route.py

RUN yum -y update && \
  yum --enablerepo=extras install -y epel-release

RUN yum install -y \
   python2-pip \
   python-devel

RUN pip install --upgrade --user setuptools pip

COPY vahter /opt/vahter
COPY setup.py /opt/

RUN pip install -r /opt/vahter/requirements.txt

WORKDIR /opt
RUN python setup.py develop

EXPOSE 5000
ENTRYPOINT ["flask", "run"]