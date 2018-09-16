import redis

# REDIS
REDIS_HOSTNAME = "redis"
REDIS_PORT = 6379
EXPIRATION = 14 * 24 * 3600  # 2 weeks


class RedisClient(object):
    def __init__(self):
        self.conn = redis.StrictRedis(host=REDIS_HOSTNAME, port=REDIS_PORT)

    def set(self, k, v, ex=EXPIRATION):
        self.conn.set(k, v, ex)

    def get(self, k):
        return self.conn.get(k)

