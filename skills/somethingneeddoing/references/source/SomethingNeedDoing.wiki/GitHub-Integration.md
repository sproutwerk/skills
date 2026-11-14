GitHub integration comes in two ways: as a means of having an auto updating macro, and as a means of having a remote dependency.

To convert any macro into a \"Git Macro\", i.e. one that is self updating, simply add a GitHub link to a file in the `Git Settings` section of the macro's metadata. Git Macros are capable of self updating (on startup), manually checking for updates (from the editor), as well as going between versions via the version history browser.

For a macro to depend on a remote file, simply add it in the `Macro Dependencies` section of the metadata as a remote dependency.