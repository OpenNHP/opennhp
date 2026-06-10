import fs from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const here = dirname(fileURLToPath(import.meta.url));
const { NHPAgent } = await import(join(here, '../../../endpoints/js-agent/dist/index.js'));

const agentPriv = fs.readFileSync(join(here, 'agent_priv.txt'),'utf8').trim();
const SERVER_C2_PUB = "SKAApHxZRTa3EPF1nlRi38neCT1H8dcJUJzcq1tUvsaIXpUj/r4DU4cZB8ApsAm9C1RGu1ZcXxm7C8frYc26+A==";

const agent = new NHPAgent({
  cipherScheme: 'gmsm',
  transport: 'relay',
  relayUrl: 'http://127.0.0.1:8081/relay',
  privateKey: agentPriv,
  logLevel: 'debug',
});
await agent.init();
agent.setIdentity({ userId: 'c2-test', deviceId: 'c2-node', organizationId: 'opennhp.org' });
agent.addServer({ publicKey: SERVER_C2_PUB });

try {
  const res = await agent.knockResource({
    serviceId: 'example',
    resourceId: 'demo',
    serverPublicKey: SERVER_C2_PUB,
  });
  console.log('KNOCK RESULT:', JSON.stringify(res, null, 2));
} catch (e) {
  console.log('KNOCK ERROR:', e.message || e);
}
