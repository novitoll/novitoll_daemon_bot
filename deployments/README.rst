NB
=====

We should migrate to normal Kubernetes (k8s) infrastructure, not this Bash and Docker-compose stuff. 
But it requires some fundings. So this will be migrated a bit later.

New chat
=====

For now, every new chat is added via Bash provisioning::

	$ ./newchat.sh "My chat" 123456

AWS system arch
=====

.. image:: docs/aws_arch.jpg

Inside AWS EC2
=======
.. image:: docs/router.jpg

