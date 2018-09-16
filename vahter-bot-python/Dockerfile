FROM centos

RUN yum -y update && \
  yum --enablerepo=extras install -y epel-release

RUN yum install -y \
   python2-pip \
   python-devel

RUN pip install --upgrade --user setuptools pip

COPY requirements.txt /tmp/
RUN pip install -r /tmp/requirements.txt

COPY vahter /opt/vahter
COPY setup.py /opt/

RUN cd /opt && python setup.py develop

WORKDIR /opt
EXPOSE 5002
ENTRYPOINT ["flask", "run"]