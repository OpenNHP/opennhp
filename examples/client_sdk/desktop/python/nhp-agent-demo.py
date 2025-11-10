import ctypes
from time import sleep

# Windows
nhp_agent = ctypes.CDLL('nhp-agent.dll')
# Linux
# mylib = ctypes.CDLL('./nhp-agent.so')
# macOS
# mylib = ctypes.CDLL('./nhp-agent.dylib')

nhp_agent.nhp_agent_init.argtypes = [ctypes.c_char_p, ctypes.c_int]
nhp_agent.nhp_agent_init.restype = ctypes.c_bool

nhp_agent.nhp_agent_init.restype = ctypes.c_int



if __name__ == '__main__':
    flag = nhp_agent.nhp_agent_init(ctypes.c_char_p(b"D:\\nhpagent"),3)
    if flag:
        print("nhp-agent init success")
    else:
        print("nhp-agent init failed")
    # start the loop knocking thread
    status = nhp_agent.nhp_agent_knockloop_start()
    if status >= 0:
        print("nhp-agent knockloop success")
        # Delay between calls
        sleep(30)
    else:
        print("nhp-agent knockloop failed")

    # stop nhp_agent
    nhp_agent.nhp_agent_close()