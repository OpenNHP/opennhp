import { NHPAgent } from '/Users/f1/project/opennhp/endpoints/js-agent/dist/index.js';
import fs from 'fs';
const agent = new NHPAgent({ cipherScheme:'gmsm', transport:'relay', relayUrl:'http://127.0.0.1:8081/relay', privateKey: fs.readFileSync('/Users/f1/project/opennhp/docker/multicluster/knock-tests/c1_priv.txt','utf8').trim(), logLevel:'error' });
await agent.init();
agent.setIdentity({ userId:'c1-test', deviceId:'c1-node', organizationId:'opennhp.org' });
const C1PUB="4/p0mIknwmVIMocRLQKil7xIthgEdZNncv9UagiBaK2kpcH7i4hEtZjpcHox+Bn7xdV+rBKNbKlV9ye6V1VCLA==";
agent.addServer({ publicKey: C1PUB });
const res = await agent.knockResource({ resourceId:'demo', serviceId:'example', serverPublicKey: C1PUB });
console.log('C1 KNOCK:', JSON.stringify(res));
