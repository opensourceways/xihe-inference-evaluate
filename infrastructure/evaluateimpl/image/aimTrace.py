# encoding: utf-8

import os
import re
import argparse

from aim import Run
from obsHandler import OBSHandler


def parse_args():
    # 创建解析
    parser = argparse.ArgumentParser(description="aim track",
                                     formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    # obs相关参数
    parser.add_argument('--obs_ak', nargs='?', type=str, default=os.environ.get('OBS_AK'), help="OBS AK")
    parser.add_argument('--obs_sk', nargs='?', type=str, default=os.environ.get('OBS_SK'), help="OBS SK")
    parser.add_argument('--obs_bucketname', nargs='?', type=str, default=os.environ.get('OBS_BUCKETNAME'),
                        help="OBS bucket name")
    parser.add_argument('--obs_endpoint', nargs='?', type=str, default=os.environ.get('OBS_ENDPOINT'),
                        help="OBS endpoint")   
    # 添加aim相关参数
    parser.add_argument('--log_path', type=str, help='log path')
    parser.add_argument('--repo', type=str, default='db/', help='storage path')
    # 解析参数
    parser.add_argument('--learning_rate_scope', type=str, default='[]', help='learning rate scope')
    parser.add_argument('--batch_size_scope', type=str, default='[]', help='batch size scope')
    parser.add_argument('--momentum_scope', type=str, default='[]',  help='momentum scope')
    # 解析参数
    args_opt = parser.parse_args()
    return args_opt

def getLogByPath(obs, log_path):
    """
    从obs中获取日志内容
    
    return: str
    """
    # 从obs中二进制读取日志文件
    res_read = obs.readFile(log_path)
    if res_read["status"] == -1:
        return {
            'status' : -1,
            'msg' : res_read['msg']
        }
    
    content = res_read['content']
    obs.close_obs()
    return {
        "status" : 200,
        "msg" : "查询成功",
        "content" : content
    }

def matchLog(log):
    """
    目前支持LossMonitor和ValAccMonitor，正则需优化
    """
    regex = r"Epoch:\[([\s|\d]+)/([\s\d]+)\].*step:\[([\s|\d]+)/([\s\d]+)\].*loss:\[([1-9]\d*\.\d*|0\.\d*[1-9]\d*|0?\.0+|0)/([1-9]\d*\.\d*|0\.\d*[1-9]\d*|0?\.0+|0)\].*time:([1-9]\d*\.\d*|0\.\d*[1-9]\d*|0?\.0+|0).*ms.*lr:([1-9]\d*\.\d*|0\.\d*[1-9]\d*|0?\.0+|0)"
    pattern = re.compile(regex) 
    res_list = pattern.findall(log)
    if len(res_list) != 0:
        return res_list, "LossMonitor"
    regex = r"Epoch:\s*\[([\s*|\d]+)/([\s*|\d]+)\],\s*Train\sLoss:\s*\[([1-9]\d*\.\d*|0\.\d*[1-9]\d*)\],\s*Accuracy:\s*([1-9]\d*\.\d*|0\.\d*[1-9]\d*)"
    pattern = re.compile(regex) 
    res_list = pattern.findall(log)
    if len(res_list) != 0:
        return res_list, "ValAccMonitor"
    return res_list, "LossMonitor"

def transData2Dict(line, kind="LossMonitor"):
    """
    将元组转换为dict
    """
    if kind == "LossMonitor":
        return {
            'epoch' : line[0],
            'step' : line[2],
            'loss' : line[5],
            'time' : line[6],
            'lr' : line[7],
            'momentum' : "-",
            'batch_size' :"-"
        }
    if kind == "ValAccMonitor":
        return {
            'epoch' : line[0],
            'loss' : line[2],
            'acc' : line[3],
            'lr' : "-",
            'momentum' :"-",
            'batch_size' : "-"
        }

def getTraceLog(res_list, kind):
    # 多组实验下的数据列表
    aimlog_list = []
    # 一种超参数下的数据列表
    aimlog_unit = []
    for i in range(len(res_list)):
        aimlog_unit.append(transData2Dict(res_list[i], kind))
        try:
            res_list[i+1]
            if int(res_list[i+1][0]) < int(res_list[i][0]):
                aimlog_list.append(aimlog_unit)
                aimlog_unit = []
        except:
            aimlog_list.append(aimlog_unit)
            break
    return aimlog_list
            
def groupTraceLog(aimlog_list, batch_size_choice, learning_rate_choice, momentum_choice):
    """
    根据用户输入的metrics补齐momentum和batch_size字段，有bug，需修改
    """
    curpos = 0
    if len(batch_size_choice) != 0:
        for i in range(len(batch_size_choice)):
            if len(learning_rate_choice) != 0:
                for j in range(len(learning_rate_choice)):
                    if len(momentum_choice) != 0:
                        for k in range(len(momentum_choice)):
                            bs = batch_size_choice[i]
                            lr = learning_rate_choice[j]
                            mom = momentum_choice[k]
                            print(bs, lr, mom)
                            for x in range(len(aimlog_list[curpos])):
                                aimlog_list[curpos][x]['lr'] = lr
                                aimlog_list[curpos][x]['batch_size'] = bs
                                aimlog_list[curpos][x]['momentum'] = mom
                            curpos += 1
                    else:
                        bs = batch_size_choice[i]
                        lr = learning_rate_choice[j]
                        mom = "-"
                        print(bs, lr, mom)
                        for x in range(len(aimlog_list[curpos])):
                            aimlog_list[curpos][x]['lr'] = lr
                            aimlog_list[curpos][x]['batch_size'] = bs
                            aimlog_list[curpos][x]['momentum'] = mom
                        curpos += 1
            else:
                if len(momentum_choice) != 0:
                    for k in range(len(momentum_choice)):
                        bs = batch_size_choice[i]
                        lr = "-"
                        mom = momentum_choice[k]
                        print(bs, lr, mom)
                        for x in range(len(aimlog_list[curpos])):
                            aimlog_list[curpos][x]['lr'] = lr
                            aimlog_list[curpos][x]['batch_size'] = bs
                            aimlog_list[curpos][x]['momentum'] = mom
                        curpos += 1
                else:
                    bs = batch_size_choice[i]
                    lr = "-"
                    mom = "-"
                    print(bs, lr, mom)
                    for x in range(len(aimlog_list[curpos])):
                        aimlog_list[curpos][x]['lr'] = lr
                        aimlog_list[curpos][x]['batch_size'] = bs
                        aimlog_list[curpos][x]['momentum'] = mom
                    curpos += 1
    else:
        if len(learning_rate_choice) != 0:
            for j in range(len(learning_rate_choice)):
                if len(momentum_choice) != 0:
                    for k in range(len(momentum_choice)):
                        bs = "-"
                        lr = learning_rate_choice[j]
                        mom = momentum_choice[k]
                        print(bs, lr, mom)
                        for x in range(len(aimlog_list[curpos])):
                            aimlog_list[curpos][x]['lr'] = lr
                            aimlog_list[curpos][x]['batch_size'] = bs
                            aimlog_list[curpos][x]['momentum'] = mom
                        curpos += 1
                else:
                    bs = "-"
                    lr = learning_rate_choice[j]
                    mom = "-"
                    print(bs, lr, mom)
                    for x in range(len(aimlog_list[curpos])):
                        aimlog_list[curpos][x]['lr'] = lr
                        aimlog_list[curpos][x]['batch_size'] = bs
                        aimlog_list[curpos][x]['momentum'] = mom
                    curpos += 1
        else:
            if len(momentum_choice) != 0:
                for k in range(len(momentum_choice)):
                    bs = "-"
                    lr = "-"
                    mom = momentum_choice[k]
                    print(bs, lr, mom)
                    for x in range(len(aimlog_list[curpos])):
                        aimlog_list[curpos][x]['lr'] = lr
                        aimlog_list[curpos][x]['batch_size'] = bs
                        aimlog_list[curpos][x]['momentum'] = mom
                    curpos += 1
            else:
                bs = "-"
                lr = "-"
                mom = "-"
                print(bs, lr, mom)
                for x in range(len(aimlog_list[curpos])):
                    aimlog_list[curpos][x]['lr'] = lr
                    aimlog_list[curpos][x]['batch_size'] = bs
                    aimlog_list[curpos][x]['momentum'] = mom
                curpos += 1
    return aimlog_list
                
def runAim(records, repo, learning_rate, momentum, batch_size, kind="LossMonitor"):
    """
    一种超参数下日志数据所能画的图
    """
    aim_run = Run(experiment = "-".join([str(learning_rate), str(momentum), str(batch_size)]), repo = repo)  # replace example IP with your tracking server IP/hostname
    # Log run parameters
    # aim_run['params'] = {
    #     'learning_rate': learning_rate,
    #     'momentum' :  momentum,
    #     'batch_size': batch_size,
    # }
    aim_run['learning_rate'] = learning_rate
    aim_run['momentum'] = momentum
    aim_run['batch_size'] = batch_size
    
    if kind == "LossMonitor":
        avgtime = 0
        for x in records:
            # 横坐标为epoch，纵坐标为loss
            aim_run.track(float(x['loss']), name='loss', epoch=x['epoch'], context={ "subset":"train" })
            # 横坐标为epoch，纵坐标为一次epoch耗费的时间
            aim_run.track(float(x['time']), name='time', epoch=x['epoch'], context={ "subset":"train" })
            avgtime += float(x['time'])
            # 横坐标为epoch，纵坐标为学习率
            aim_run.track(float(x['lr']), name='learning_rate', epoch=x['epoch'], context={ "subset":"train" })
        avgtime = avgtime / len(records)
        for x in records:
            # 横坐标为epoch, 纵坐标为
            aim_run.track(avgtime, name='average_time', epoch=x['epoch'], context={ "subset":"train" })
    if kind == "ValAccMonitor":
        for x in records:
            # 横坐标为epoch，纵坐标为loss
            aim_run.track(float(x['loss']), name='loss', epoch=x['epoch'], context={ "subset":"train" })
            # 横坐标为epoch，纵坐标为一次epoch耗费的时间
            aim_run.track(float(x['acc']), name='time', epoch=x['epoch'], context={ "subset":"test" })

def main(args):  
    # 读取obs相关超参数
    obs_ak = args.obs_ak
    obs_sk = args.obs_sk
    obs_bucketname = args.obs_bucketname
    obs_endpoint = args.obs_endpoint
    
    # 读取aim相关超参数
    log_path = args.log_path
    repo = args.repo
    # user = args.user
    # project = args.project
    # job_id = args.job_id
    learning_rate_scope = eval(args.learning_rate_scope)
    batch_size_scope = eval(args.batch_size_scope)
    momentum_scope = eval(args.momentum_scope)
    print("this is matrics!!", learning_rate_scope, batch_size_scope, momentum_scope)
   
    obs = OBSHandler(obs_ak, obs_sk, obs_bucketname, obs_endpoint)
    # 通过user,project,job_id获取日志文件并读取内容
    get_res = getLogByPath(obs, log_path)
    if get_res['status'] == -1:
        return {
            'status' : -1,
            'msg' : get_res['msg']
        }
    log_content = get_res['content']
    
    # 根据metrics（lr、bs、momentum）对日志内容进行处理，分对比组
    # 匹配LossMonitor日志内容
    res_list, kind = matchLog(log_content)
    print("this is res_list & kind", res_list, kind)
    # 得到跟踪数据雷彪
    aimlog_list = getTraceLog(res_list, kind)
    # 根绝metrics分组，补齐数据
    aimlog_list = groupTraceLog(aimlog_list, batch_size_scope, learning_rate_scope, momentum_scope)
    
    # 启动aim，保存数据
    records = aimlog_list
    for record in records:
        lr = record[0]['lr']
        mom = record[0]['momentum']
        bs = record[0]['batch_size']
        # print(lr, mom, bs)
        runAim(record, repo, lr, mom, bs, kind)
    
                
                
if __name__ == "__main__":
    # batch_size_choice = [32, 64, 128]
    # learning_rate_choice = [0.1, 0.01, 0.02, 0.04]
    # momentum_choice = [0.9, 0.99]    
    # 获取超参数
    args_opt = parse_args()
    # 使用aim跟踪日志数据
    main(args_opt)
