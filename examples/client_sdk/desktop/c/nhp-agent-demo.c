#include <stdio.h>
#include <unistd.h>
#include "nhp-agent.h"

int main() {
    // Initialize nhp_agent, only one nhp_agent singleton is allowed per process.
    nhp_agent_init(".", 3);

    // Set the user information for the knock-on-the-door feature.
    nhp_agent_set_knock_user("zengl", NULL, NULL, NULL);

    // Set NHP server information.
    // If there is already a server.toml configured under ./etc/, the call to
    // nhp_agent_add_server can be omitted — clusters defined in server.toml are
    // already registered at init time. Timestamp date is at https://unixtime.org/
    const char *server_pubkey = "replace_with_actual_publickeybase64";
    nhp_agent_add_server(server_pubkey, "192.168.1.66", NULL, 62206, 1748908471);

    // Resources are now bound to an nhp-server CLUSTER by name. The cluster
    // identifier is:
    //   - For SDK callers using nhp_agent_add_server: the server's pubkey
    //     base64 (single-instance clusters are auto-named after their pubkey).
    //   - For callers loading server.toml: the [[Servers]] Name field.
    // The legacy (serverIp, serverHostname, serverPort) trailing parameters
    // were removed in v1.x — they were silently ignored when ServerPubKey was
    // set, which let resource.toml display addresses the agent never dialed.
    char *ret = nhp_agent_knock_resource("example", "demo", server_pubkey);
    printf("knock return: %s\n", ret);
    nhp_free_cstring(ret);

    // Immediately close the agent's access to the example/demo resources;
    // otherwise access auto-closes after the door-opening duration expires.
    nhp_agent_exit_resource("example", "demo", server_pubkey);

    // Turn off and release nhp_agent.
    nhp_agent_close();
    return 0;
}

