import { ConnectorError, StandardCommand } from '@sailpoint/connector-sdk'
import { MyClient } from './my-client'

const mockConfig: any = {
    token: 'xxx123'
}

describe('connector client unit tests', () => {

    const myClient = new MyClient(mockConfig)

    it('connector client list accounts', async () => {
        let allAccounts = await myClient.getAllAccounts()
        expect(allAccounts.length).toStrictEqual(2)
    })

    it('connector client get account', async () => {
        let account = await myClient.getAccount('john.doe')
        expect(account.username).toStrictEqual('john.doe')
    })

    it('connector client test connection', async () => {
        expect(await myClient.testConnection()).toStrictEqual({})
    })

    it('connector client test connection', async () => {
        expect(await myClient.testConnection()).toStrictEqual({})
    })

    it('invalid connector client', async () => {
        try {
            new MyClient({})
        } catch (e) {
            expect(e instanceof ConnectorError).toBeTruthy()
        }
    })
})
