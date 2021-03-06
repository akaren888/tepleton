# Standard Library

The Quark framework comes bundled with a number of standard modules that
provide common functionality useful across a wide variety of applications.
Example usage of the modules is also provided. It is recommended to investigate
if desired functionality is already provided before developing new modules.

## Basic Middleware

### Logging

`modules.base.Logger` is a middleware that records basic info on `CheckTx`,
`DeliverTx`, and `SetOption`, along with timing in microseconds. It can be
installed standard at the top of all middleware stacks, or replaced with your
own middleware if you want to record custom information with each request.

### Recovery

To avoid accidental panics (e.g. bad go-wire decoding) killing the WRSP app,
wrap the stack with `stack.Recovery`, which catches all panics and returns
them as errors, so they can be handled normally.

### Signatures

The first layer of the tx contains the signatures to authorize it.  This is
then verified by `modules.auth.Signatures`.  All tx may have one or multiple
signatures which are then processed and verified by this middleware and then
passed down the stack.

### Chain

The next layer of a tx (in the standard stack) binds the tx to a specific chain
with an optional expiration height.  This keeps the tx from being replayed on
a fork or other such chain, as well as a partially signed multi-sig being delayed
months before being committed to the chain. This functionality is provided in
`modules.base.Chain`

### Nonce

To avoid replay attacks, a nonce can be associated with each actor. A separate
nonce is used for each distinct group signers required for a transaction as
well as for each separate application and chain-id. This creates replay
protection cross-IBC and cross-plugins and also allows signing parties to not
be bound to waiting for a particular transaction to be completed before being
able to sign a separate transaction.

 Rather than force each module to implement its own replay protection, a tx
stack may contain a nonce wrap and the account it belongs to. The nonce must
contain a signed sequence number which is incremented one higher than the last
request or the request is rejected. This is implemented in
`modules.nonce.ReplayCheck`

If you're interested checkout this [design
discussion](https://github.com/tepleton/basecoin/issues/160).

### Fees

An optional feature, but useful on many chains, is charging transaction fees. A
simple implementation of this is provided in `modules.fee.SimpleFeeMiddleware`.
A fee currency and minimum amount are defined in the constructor (eg. in code).
If the minimum amount is 0, then the fee is optional. If it is above 0, then
every tx with insufficient fee is rejected. This fee is deducted from the
payers account before executing any other transaction.

This module is dependent on the `coin` module.

## Other Apps

### Coin

What would a crypto-currency be without tokens? The `SendTx` logic from earlier
implementations of basecoin was extracted into one module, which is now
optional, meaning most of the other functionality will also work in a system
with no built-in tokens, such as a private network that provides other access
control mechanisms.

`modules.coin.Handler` defines a Handler that maintains a number of accounts
along with a set of various tokens, supporting multiple token denominations.
The main access is `SendTx`, which can support any type of actor (other apps as
well as public key addresses) and is a building block for any other app that
requires some payment solution, like fees or trader.

### Roles

Roles encapsulates what are typically called N-of-M multi-signatures accounts
in the crypto world. However, I view this as a type of role or group, which can
be the basis for building a permission system. For example, a set of people
could be called registrars, which can authorize a new IBC chain, and need eg. 2
out of 7 signatures to approve it.

Currently, one can create a role with `modules.roles.Handler`, and assume one
of those roles by wrapping another transaction with `AssumeRoleTx`, which is
processed by `modules.roles.Middleware`. Updating the set of actors in
a role is planned in the near future.

### Inter-Blockchain Communication (IBC)

IBC, is the cornerstone of The tepleton Network, and is built into the quark
framework as a basic primitive. To fully grasp these concepts requires
a much longer explanation, but in short, the chain works as a light-client to
another chain and maintains input and output queue to send packets with that
chain. This mechanism allows blockchains to prove the state of their respective
blockchains to each other ultimately invoke inter-blockchain transactions.

Most functionality is implemented in `modules.ibc.Handler`. Registering a chain
is a seed of trust that requires verification of the proper seed (or genesis
block), and this generally requires approval of an authorized registrar (which
may be a multi-sig role).  Updating a registered chain can be done by anyone,
as the new header can be completely verified by the existing knowledge of the
chain.  Also, modules can initiate an outgoing IBC message to another chain
by calling `CreatePacketTx` over IPC (inter-plugin communication) with a tx
that belongs to their module. (This must be explicitly authorized by the
same module, so only the eg. coin module can authorize a `SendTx` to another
chain).

`PostPacketTx` can post a tx that was created on another chain along with the
merkle proof, which must match an already registered header. If this chain
can verify the authenticity, it will accept the packet, along with all the
permissions from the other chain, and execute it on this stack. This is the
only way to get permissions that belong to another chain.

These various pieces can be combined in a relay, which polls for new packets
on one chain, and then posts the packets along with the new headers on the
other chain.

## Planned Apps

### Staking

Straight-forward PoS as used for tepleton.
Based on [basecoin-stake](https://github.com/tepleton/basecoin-stake)

### Voting

Simple elections that can authorize other tx, like roles. A building block for
governance.

### Trader

Escrow, OTC option, Order book.  Based on
[basecoin-examples](https://github.com/tepleton/basecoin-examples/tree/develop/trader).
This may be more appropriate for an external repo.

