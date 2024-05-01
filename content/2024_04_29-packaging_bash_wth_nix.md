
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

##### __Note:__ If you want to skip the article and just see the code, check it out [here](https://github.com/cesarFuhr/notes-script).

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
  rg --pretty '\[ \]' . && \ # Scan with `rg`.
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
  rg --pretty '\[x\]' . && \ # Scan for completed checkboxes.
  popd > /dev/null # Move back to the callers directory.
```

And that's it! That is all I needed... But I really wanted to integrate this seamlessly with OS and package it in a way I could bundle all the dependencies with it.

## Nix for the rescue

I am using NixOS as my daily driver for more than two years now, but never had packaged anything but dev shells with it yet. This was a great opportunity to exercise my Nix brain cells and also dive a little deeper into a functional language.

Nix is a functional declarative language developed to tackle the software building problem (its also a package manager and build tool). Although I am used to run and manage my NixOS configuration and some development shells, actually packaging a piece of software was a new thing to me. Being the kind of person I am, I imposed two extra restrictions on myself: 

- Only use pure nix and the Nix [builtins](https://nixos.org/manual/nix/stable/language/builtins.html), don't rely on any external packages.
- Use the flake experimental (but pretty much stable) feature.

Let's start by building the binaries, which is a our smallest scope, then increase the scope to arrive at the final Nix package. To do that we need to define a function to create the individual package.

```nix
pack = ({ packageName, buildInputs }:
  let
    p = import nixpkgs { system = system; };

    script = (p.writeScriptBin packageName (builtins.readFile ./${packageName}.sh))
        .overrideAttrs (old: { buildCommand = "${old.buildCommand}\n patchShebangs $out"; });
  in
  p.symlinkJoin {
    name = packageName;
    paths = [ script ] ++ buildInputs;
    buildInputs = [ p.makeWrapper ];
    postBuild = "wrapProgram $out/bin/${packageName} --prefix PATH : $out/bin";
  });
```

This little snippet of code will create a Nix package based on a `packageName` and the packages `buildInputs`. It also has an indirect dependency with `nixpkgs`, which by itself is dependant on the `system` variable, but lets not focus on that right now. The following two lines are what is actually building the script into an executable, linking and wrapping the bash script into an executable in the nix store `bin` directory. All the `p.{function_name}` calls are nix packages available (something like a standard library, you can check their docs [here](https://nixos.org/manual/nixpkgs/stable/)) through `nixpkgs`.

```nix
    script = (p.writeScriptBin packageName (builtins.readFile ./${packageName}.sh))
        .overrideAttrs (old: { buildCommand = "${old.buildCommand}\n patchShebangs $out"; });
```

Then, since we are building such a simple application, what we need to do is just package it. `p.symlinkJoin` will gather all the build inputs (the script dependencies) and the script itself and create a single folder with all the links. Then we use `wrapProgram` function to make sure the app will be available in your PATH after installation.

With the packaging function defined we can build the system specific copy of our notes app and inject them in the final derivation. Nix expects the result of the evaluation to be an [attribute set](https://nixos.org/manual/nix/stable/language/values#attribute-set), where for every supported [system](https://nixos.org/manual/nix/stable/language/derivations#attr-system) there should be a [derivation](https://nixos.org/manual/nix/stable/language/derivations) of the app. This is where the `builtins` come in handy, they help us with mapping the data into different formats to achieve the needed result spec.

`builtins.map` receives a mapping function and a [list](https://nixos.org/manual/nix/stable/language/values#list) and applies that function to every element, returning the resulting list of transformed elements. `builtins.listToAttrs` on the other hand receives a list of attribute sets with `name` and `value` attributes and reduces them into an attribute set, where the keys will be the `name`s and the values will the `value`s of the list elements.

The first thing we do, is define the list of systems we will support. Then we define the output attribute set with a single attribute `packages`. Each key in the `packages` set should be a system and its value an set of built packages.

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

              pack = ({ packageName, buildInputs }: /* We covered this function already. */ });
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
