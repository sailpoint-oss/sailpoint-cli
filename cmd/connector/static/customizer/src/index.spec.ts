import { connectorCustomizer } from './index'
import { CustomizerType, StandardCommand, AssumeAwsRoleRequest, AssumeAwsRoleResponse } from '@sailpoint/connector-sdk'

const mockConfig: any = {
    token: 'xxx123',
}
process.env.CONNECTOR_CONFIG = Buffer.from(JSON.stringify(mockConfig)).toString('base64')

describe('connector customizer unit tests', () => {
    it('should not change input from beforeStdAccountReadHandler', async () => {
        let customizer = await connectorCustomizer()
        let input = {
            identity: 'john.doe',
        }
        let updatedInput = await customizer._exec(
            customizer.handlerKey(CustomizerType.Before, StandardCommand.StdAccountRead),
            {
                reloadConfig() {
                    return Promise.resolve()
                },
                assumeAwsRole(assumeAwsRoleRequest: AssumeAwsRoleRequest): Promise<AssumeAwsRoleResponse> {
                    return Promise.resolve(
                        new AssumeAwsRoleResponse('ccessKeyId', 'secretAccessKey', 'sessionToken', '123')
                    )
                },
            },
            input
        )

        expect(input).toStrictEqual(updatedInput)
    })

    it('should add location attribute from afterStdAccountReadHandler', async () => {
        let customizer = await connectorCustomizer()
        let output = await customizer._exec(
            customizer.handlerKey(CustomizerType.After, StandardCommand.StdAccountRead),
            {
                reloadConfig() {
                    return Promise.resolve()
                },
                assumeAwsRole(assumeAwsRoleRequest: AssumeAwsRoleRequest): Promise<AssumeAwsRoleResponse> {
                    return Promise.resolve(
                        new AssumeAwsRoleResponse('ccessKeyId', 'secretAccessKey', 'sessionToken', '123')
                    )
                },
            },
            {
                identity: '',
                attributes: {
                    username: 'john.doe',
                    firstName: 'john',
                    lastName: 'doe',
                    email: 'john.doe@example.com',
                },
            }
        )

        expect(output.attributes.location).toStrictEqual('Austin')
    })
})
