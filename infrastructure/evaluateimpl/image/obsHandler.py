# encoding: utf-8

import time
import argparse
import traceback

from obs import ObsClient


class OBSHandler:
    def __init__(self, obs_ak, obs_sk, obs_bucketname, obs_endpoint):
        self.access_key = obs_ak
        self.secret_key = obs_sk
        self.bucket_name = obs_bucketname
        self.endpoint = obs_endpoint
        self.server = "https://" + obs_endpoint
        self.obsClient = self.init_obs()
        self.maxkeys = 100  #查询的对象最大个数

    # 初始化obs
    def init_obs(self):
        obsClient = ObsClient(
            access_key_id = self.access_key,
            secret_access_key = self.secret_key,
            server = self.server
        )
        return obsClient

    def close_obs(self):
        self.obsClient.close()
      
    def readFile(self, path):
        """
        二进制读取文本文件
        :param configPath: 为当前仓库下的相对路径，比如train/config1.json
        :return:
        """
        try:
            resp = self.obsClient.getObject(self.bucket_name, path, loadStreamInMemory=True)
            if resp.status < 300:
                # 获取对象内容
                return {
                    "status": 200,
                    "msg": "获取文件成功",
                    "content": bytes.decode(resp.body.buffer, "utf-8"),
                    "size": resp.body.size
                }
            else:
                return {
                    "status": -1,
                    "msg": "获取失败，失败码: %s\t 失败消息: %s" % (resp.errorCode, resp.errorMessage),
                    "content": "",
                    "size": 0
                }
        except:
            return {
                    "status": -1,
                    "msg": "获取失败, obs服务器挂了",
                }
