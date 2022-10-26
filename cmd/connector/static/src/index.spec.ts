import { connector } from './index'
import { StandardCommand } from '@sailpoint/connector-sdk'
import { PassThrough } from 'stream'

const mockConfig: any = {
    token: 'xxx123'
}
process.env.CONNECTOR_CONFIG = Buffer.from(JSON.stringify(mockConfig)).toString('base64')

describe('connector unit tests', () => {

    it('connector SDK major version should be 0', async () => {
        expect((await connector()).sdkVersion).toStrictEqual(0)
    })

    it('should execute stdTestConnectionHandler', async () => {
        await (await connector())._exec(
            StandardCommand.StdTestConnection,
            {},
            undefined,
            new PassThrough({ objectMode: true }).on('data', (chunk) => expect(chunk).toStrictEqual({}))
        )
    })
})
