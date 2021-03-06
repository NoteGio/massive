Massive is a set of command-line tools for creating command-line scripts for
interacting with the Ethereum blockchain in general, and 0x in particular.

Most Massive commands will have a source, a sink, or both. Sources can either
be files or stdin, while sinks can either be files or stdout. Some commands
will have other attributes.

An example pipeline:

    # Read records out of a CSV
    massive 0x csv --input transactions.csv | \
    # Get fees from the relayer
    massive 0x getFees --target https://api.openrelay.xyz | \
    # Add the current timestamp as a nonce
    massive 0x timestampSalt | \
    # Set the expiration date for 10 days in the future
    massive 0x expiration --duration 864000 | \
    # Sign with the provided key
    massive 0x sign $KEY_FILE | \
    # Verify that the transaction is fillable
    massive 0x verify | \
    # Upload the transaction to a 0x relayer
    massive 0x upload --target https://api.openrelay.xyz

This would get us from a CSV containing a list of tokens to be traded to
uploaded 0x orders.
