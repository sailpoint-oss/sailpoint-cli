import { connector } from './index'
import { Connector, RawResponse, ResponseType, StandardCommand, AssumeAwsRoleRequest, AssumeAwsRoleResponse } from '@sailpoint/connector-sdk'
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
            {reloadConfig() {
                return Promise.resolve()
            },
            assumeAwsRole(assumeAwsRoleRequest: AssumeAwsRoleRequest): Promise<AssumeAwsRoleResponse> {
                return Promise.resolve(new AssumeAwsRoleResponse('accessKeyId', 'secretAccessKey', 'sessionToken', "123"))
            }
        },
            undefined,
            new PassThrough({ objectMode: true }).on('data', (chunk) => expect(chunk).toStrictEqual(new RawResponse ({}, ResponseType.Output)))
        )
    })
})
