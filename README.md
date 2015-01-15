# larasync
larasync is an end-to-end encrypted, simple and fast self-hosted file synchronization solution.

Use it to synchronize all your documents across your devices (no smartphones yet, sorry) with the help of a (not necessarily trusted) machine acting as your server.

## Status
larasync is still **alpha** software.
Not even the developers use it for real data yet.

A git-like command line interface is the only available interface at the moment.
There are no signed releases, there is no conflict resolution, no automatic file synchronization, no kind of API or on-disk format stability guarantees are made, there are many other pitfalls and there *will* be bugs, so:
**DO NOT USE IT IN PRUDUCTION YET**

## Platforms
Development primarly happens on Linux and Mac OS X at the moment, but sporadic experiments on Windows seem to be successful as well.

## Usage
1. Environment setup (required on both client and your server)
  - Install go ([Official docs](https://golang.org/doc/install))
  - Get larasync:
    `go get github.com/hoffie/larasync/cmd/lara`

2. Generate an admin secret (on the client)  
   This will be used by all users of your system to register new repositories with the server.
   To do this, run `lara admin-secret` and choose a passphrase.
   Remember the resulting hash for the next step.

3. Configure the server
   - On the server, create a config file based on the [example configuration](doc/larasync-server.gcfg.example) (rename it to `larasync-server.gcfg`).
   - Change the *listen* address as you wish (running it on an external interface sounds like a good idea)
  - Change *adminpubkey* to the value you got from `lara admin-secret`.
  - Set *basepath* to an existing directory where all your repositories should be stored.
  - Start the server by running `lara server` in the directory containing the config file.

4. Create a new repository (on your first client)
   - `lara init my-repository` will create the sub-directory `my-repository`; change to it using `cd my-repository`
   - Register it with the server using `lara register HOST:PORT my-repository`; You will be asked to enter the *admin secret* chosen during setup.
   - Create files, documents and pictures in this repository as you wish; automatically synchronize all your local changes with the server using `lara sync`.

5. Integrate one or more other clients
   - Run `lara authorize-new-client` on your first client (or any other already set-up client).
   - Forward the resulting URL to your new client using a secure transport (GPG-encrypted mail should work).
   - On the new client, the first and only command you have to run is `lara clone URL-FROM-ABOVE my-local-repository`; with this URL and the included temporary keys, it will be provided with the necessary encryption keys to be part of the system.
   - All previously added data should already be available. As always, run `lara sync` after any changes.

Also refer to `lara help` for a full list of supported commands.

## Security
Security is a top-priority for us.
More documentation on the selected technology, threat vectors and mitigations will be published in the future (we use NaCl for encryption along with various standard algorithms for HMACs and signatures).
Should you spot any security-relevant problems, we are eager to hear from you! Please [contact us](mailto:team@larasync.org) in this case.

## License
larasync is licensed under the [AGPLv3](LICENSE.AGPLv3).

## Authors
This project is maintained by [Christoph Brand](mailto:christoph@larasync.org) ([@cbrand](https://github.com/cbrand)) and [Christian Hoffmann](mailto:hoffie@larasync.org) ([@hoffie](https://github.com/hoffie)).
