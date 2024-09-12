import { connector } from './index'
import { Connector, RawResponse, ResponseType, StandardCommand } from '@sailpoint/connector-sdk'
import { PassThrough } from 'stream'

const mockConfig: any = {
    token: 'xxx123'
}
process.env.CONNECTOR_CONFIG = Buffer.from(JSON.stringify(mockConfig)).toString('base64')

describe('connector unit tests', () => {

    it('connector SDK major version should be the same as Connector.SDK_VERSION', async () => {
        expect((await connector()).sdkVersion).toStrictEqual(Connector.SDK_VERSION)
    })

    it('should execute stdTestConnectionHandler', async () => {
        await (await connector())._exec(
            StandardCommand.StdTestConnection,
            {},
            undefined,
            new PassThrough({ objectMode: true }).on('data', (chunk) => expect(chunk).toStrictEqual(new RawResponse ({}, ResponseType.Output)))
        )
    })
})
