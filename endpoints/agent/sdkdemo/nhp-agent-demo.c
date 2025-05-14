#include <stdio.h>
#include <unistd.h>
#include "nhp-agent.h"

int main() {
    // Initialize nhp_agent, only one nhp_agent singleton is allowed per process.
    nhp_agent_init(".", 3);

    // Set the user information for the knock-on-the-door feature.
    nhp_agent_set_knock_user("zengl", NULL, NULL, NULL);

    // Set NHP server information
    // If there is already a configuration file for the server, the call to nhp_agent_add_server can be omitted
    // Timestamp date is visible at https://unixtime.org/
    nhp_agent_add_server("replace_with_actual_publickeybase64", "192.168.1.66", NULL, 62206, 1748908471);

    // Send a request to the server to access the resource example/demo, and return information in the form of a JSON format string
    // Note: The resource information here is an independent input, and is unrelated to the resource information saved in the configuration file
    char *ret = nhp_agent_knock_resource("example", "demo", "192.168.1.66", NULL, 62206);
    printf("knock return: %s\n", ret);

    // Immediately close the agent's access to the example/demo resources,
    // if not invoked, access permission will automatically close after the door opening duration has passed.
    nhp_agent_exit_resource("example", "demo", "192.168.1.66", NULL, 62206);

    // Turn off and release nhp_agent.
    nhp_agent_close();
    return 0;
}

