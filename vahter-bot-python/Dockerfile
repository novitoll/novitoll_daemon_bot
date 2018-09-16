FROM centos

RUN yum -y update && \
  yum --enablerepo=extras install -y epel-release

RUN yum install -y \
   python2-pip \
   python-devel

RUN pip install --upgrade --user setuptools pip

COPY requirements.txt /tmp/
RUN pip install -r /tmp/requirements.txt

COPY entrypoint.sh /opt/

VOLUME /opt

WORKDIR /opt
EXPOSE 5000
ENTRYPOINT ["/bin/bash", "-c", "./entrypoint.sh"]