import {
    Context,
    createConnectorCustomizer,
    readConfig,
    logger,
    StdAccountReadInput,
    StdAccountReadOutput,
    StdTestConnectionOutput,
} from '@sailpoint/connector-sdk'

// Connector customizer must be exported as module property named connectorCustomizer
export const connectorCustomizer = async () => {

    // Get connector source config
    const config = await readConfig()

    return createConnectorCustomizer()
        .afterStdTestConnection(async (context: Context, output: StdTestConnectionOutput) => {
            logger.info('Running after test connection')
            return output
        })
        .beforeStdAccountRead(async (context: Context, input: StdAccountReadInput) => {
            logger.info(`Running before account, for account ${input.identity}`)
            return input
        })
        .afterStdAccountRead(async (context: Context, output: StdAccountReadOutput) => {
            logger.info(`Running after account read to add custom attribute "location"`)

            output.attributes.location = 'Austin'
            return output
        })
}
