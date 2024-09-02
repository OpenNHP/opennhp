#include <stdio.h>
#include <unistd.h>
#include "nhp-agent.h"

int main() {
    // 初始化nhp_agent，一个进程只能有一个nhp_agent单例
    nhp_agent_init(".", 3);

    // 设置敲门的用户信息
    nhp_agent_set_knock_user("zengl", NULL, NULL, NULL);

    // 设置nhp服务器信息
    // 如果已经存在server的配置文件，nhp_agent_add_server调用可以省略
    // 时间戳日期可见 https://unixtime.org/
    nhp_agent_add_server("replace_with_actual_publickeybase64", "192.168.1.66", NULL, 62206, 1748908471);

    // 向服务器发送请求访问资源example/demo，返回信息为json格式字符串
    // 注：此处的资源信息为独立输入，与配置文件中已保存的资源信息无关
    char *ret = nhp_agent_knock_resource("example", "demo", "192.168.1.66");
    printf("knock return: %s\n", ret);

    // 立即关闭agent对example/demo资源的访问，如果不调用，则访问权限会在开门时长经过后自行关闭
    nhp_agent_exit_resource("example", "demo", "192.168.1.66");

    // 关闭并释放nhp_agent
    nhp_agent_close();
    return 0;
}

