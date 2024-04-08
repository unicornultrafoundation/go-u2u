## Basics
    Inspired by EIP712, EIP2938 and ERC4337 we have created a new standard for transaction validation and execution, it called Native Account Abstraction (NAA). The main goal of this standard is to provide a way to abstract the some of the validation and execution process from the EVM to custom contracts, allowing for more flexibility and extensibility in the way transactions are validated and executed (mainly tx.origin and alternative fee validation).
    Due to complexitiy of the EIP712 and EIP2938, we have decided to create a new standard that will be more simple and easy to understand.

## Transaction flow

### The validation step

Step 1. The validator checks that the nonce of the transaction in the right order.

Step 2. The validator calls the `IAccount.validateTransaction` method of the account. If it does not revert, proceed to the next step.

Step 3. The validator increase nonce of the `IAccount`.

Step 4.
    (no paymaster). The validator calls the `IAccount.payForTransaction`. If it does not revert, proceed to the next step.
    (paymaster). The validator calls the `IAccount.prepareForPaymaster` method of the sender. If this call does not revert, then the `IAccount.validateAndPayForPaymasterTransaction` method of the paymaster is called. If it does not revert too, proceed to the next step.

Step 5. The validator verifies that the validator has received at least tx.gasPrice * tx.gasLimit ETH to the validator. If it is the case, the verification is considered complete and we can proceed to the next step.

### The execution step

Step 6. The validator calls the executeTransaction method of the account.

Step 7. (only in case the transaction has a paymaster) The postTransaction method of the paymaster is called. This step should typically be used for refunding the sender the unused gas in case the paymaster was used to facilitate paying fees in ERC-20 tokens.

### Fee
    In EIP4337 you can see three types of gas limits: verificationGas, executionGas, preVerificationGas
    By default, calling estimateGas adds a constant to cover charging the fee and the signature verification for EOA accounts.

## Contract magic value

Now, both accounts and paymasters are required to return a certain magic value upon validation. This magic value will be
enforced to be correct on the mainnet, but will be ignored during fee estimation. Unlike Ethereum, the signature
verification + fee charging/nonce increment are not included as part of the intrinsic costs of the transaction. These
are paid as part of the execution and so they need to be estimated as part of the estimation for the transaction’s
costs.

Generally, the accounts are recommended to perform as many operations as during normal validation, but only return the
invalid magic in the end of the validation. This will allow to correctly (or at least as correctly as possible) estimate
the price for the validation of the account.
