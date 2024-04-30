
##### April 29th, 2024

# Packaging bash with Nix
#### Packaging some simple note taking scripts with Nix shouldn't be so hard right? right?

So, I have been trying to find a note taking app that suites my needs for a while. After some search, trying out some GUI based apps, with sync, without sync... I got tired of searching and overwhelmed with so many options, ended up deciding to build something simple, but something that would work for me.

I wanted my app to be:

- Fast to access
- Fast to search
- Support Markdown format
- Support some sort of task management
- Support notes name spacing (I want to be able to have work and personal notes separate)
- And (the killer feature) support vim motions

But wait a second, isn't this blog post about a couple of simple bash scripts? Yes it is!

## The fabulous "app"

Lets review the list of requirements:

> Fast to access

This means it should be local and TUI, otherwise I would need to write a service... That doesn't sound very simple.

> Fast to search

Having a bunch of local files makes it pretty easy to search with simple bash tools like `find`, `grep` or `rg`.

> Support Markdown format

Well... It's not like it needs to support markdown previews, it's more of _should not prevent me_ of using such format. So any plain text file should do it.

> Support some sort of task management

Ok, now its going to force me to write a real app isn't it? Well, no. I can probably get away with some simple convention, like `[ ]` this is something that should be done eventually and `[x]` is something already done. Basically using what Markdown already gives me.

> Support notes name spacing

This looks like a great application for my amazing ground breaking feature: File Directories!

> And support for vim motions

What I need is to use a terminal based, fast and configurable editor. Why not Neovim (my editor of choice)? I will admit it... It is kinda cheating.

Lately I am trying to get more proficient with bash and the command line, so I thought to myself: what if I just bashed (pun intended) together a couple of scripts and called it a day? 

Usually I like to aggregate my notes for a day in the same file, this keeps them organized and pretty easy to locate. I also would like to be able to know the month and year of the note, but that (to me) looks more like directories. So, putting it all together, this little script was born.


```bash
# This is the name spacing "feature".
# I am assuming personal here if no other namespace was gives 
# as the first argument.
SUBJECT="personal"
if [ -n "$1" ]; then
  SUBJECT=$1
fi

# Then I compose the file name and directory.
NOTES_DIR=$(date +"%Y/%m")
FILE=$(date +"%d")

# Make sure the directory path exists and move to it.
mkdir -p ~/.notes/$SUBJECT && pushd $_ > /dev/null && mkdir -p $NOTES_DIR

# Call the editor for the notes file.
$EDITOR "./${NOTES_DIR}/${FILE}.md"

# Move the current directory to the original caller directory.
popd > /dev/null
```

Yes, really simple, but that was the point. To navigate in the notes I use a fuzzy finder inside Neovim and I am ready to go. Notes taking part is done, lets see what we can do with the task "management" part.

My plan was to use a convention following Markdown checkboxes syntax. What I need is something that will find and show me in a simple way the tasks I have and in witch note file it lives. I also need something to show me the ones that are done.

The show me what to do script:

```bash
# Same name spacing with directories.
SUBJECT="personal"
if [ -n "$1" ]; then
  SUBJECT=$1
fi

# Move to the subject folder.
pushd ~/.notes/$SUBJECT > /dev/null && \
  rg --pretty '\[ \]' . && \ # Scan with rg.
  popd > /dev/null # Move back to the callers directory.
```

And the show what I have done script:

```bash
# Same name spacing with directories.
SUBJECT="personal"
if [ -n "$1" ]; then
  SUBJECT=$1
fi

# Move to the subject folder.
pushd ~/.notes/$SUBJECT > /dev/null && \
  rg --pretty '\[x\]' . && \
  popd > /dev/null # Move back to the callers directory.
```

And that's it! That is all I needed... But I really wanted to integrate this seamlessly with OS and package it in a way I could bundle all the dependencies with it.

## Nix for the rescue

```nix
{
  description = "A very basic note taking script";
  outputs = inputs@{ self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
    in
    {
      packages = builtins.listToAttrs
        (builtins.map
          (system:
            let
              p = import nixpkgs { system = system; };

              pack = ({ packageName, buildInputs }:
                let
                  script = (p.writeScriptBin packageName (builtins.readFile ./${packageName}.sh)).overrideAttrs (old: {
                    buildCommand = "${old.buildCommand}\n patchShebangs $out";
                  });
                in
                p.symlinkJoin {
                  name = packageName;
                  paths = [ script ] ++ buildInputs;
                  buildInputs = [ p.makeWrapper ];
                  postBuild = "wrapProgram $out/bin/${packageName} --prefix PATH : $out/bin";
                });
            in
            {
              name = system;
              value = rec {
                default = notes;
                notes = pack { packageName = "notes"; buildInputs = [ p.coreutils ]; };
                todo = pack { packageName = "todo"; buildInputs = [ p.coreutils p.ripgrep ]; };
                todo-done = pack { packageName = "todo-done"; buildInputs = [ p.coreutils p.ripgrep ]; };
              };
            })
          systems);
    };
}
```
