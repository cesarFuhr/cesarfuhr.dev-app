
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

The first thing we do, is define the list of systems we will support. We will map this list into the final distribution, making the app available to these systems.

```nix
let
  systems = [
    "x86_64-linux"
    "aarch64-linux"
    "x86_64-darwin"
    "aarch64-darwin"
  ];
in
```

Now we need to declare the building process of the app, which will expose three commands `notes`, `todo` and `todo-done`. This is done by the `builtins.map` function, that receives a mapping function and a [list](https://nixos.org/manual/nix/stable/language/values#list) and applies that function to every element, returning the resulting list of transformed elements. 

```nix
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
```

Each element inside the list returned by the `builtins.map` function is a set with the following form:

```nix
{
    name = "system-name";
    value = {
        default = notes;
        notes = notes_package;
        todo = todo_package;
        todo-done = todo-done_package;
    };
}
```

We need such arrangement to feed it into `builtins.listToAttrs`, which receives a list of attribute sets with `name` and `value` attributes and reduces them into an attribute set, where the keys will be the `name`s and the values will the `value`s of the list elements. The end result will be an set such as this:

```nix
{
    "system-0" = {
        default = notes;
        notes = notes_package;
        todo = todo_package;
        todo-done = todo-done_package;
    };
#       .   
#       .   
#       .   
    "system-N" = {
        default = notes;
        notes = notes_package;
        todo = todo_package;
        todo-done = todo-done_package;
    };
}
```

Then we define this attribute set as the `packages` attribute value in the final set. Executing `nix flake show` we will get the following output for a `x86_64-linux` machine:

```bash
$ nix flake show
git+file:///path/to/project/notes?ref=refs/heads/main&rev=fcf77dbb83cd3c22cac6a1358365d234ed20627d
└───packages
    ├───aarch64-darwin
    │   ├───default omitted (use '--all-systems' to show)
    │   ├───notes omitted (use '--all-systems' to show)
    │   ├───todo omitted (use '--all-systems' to show)
    │   └───todo-done omitted (use '--all-systems' to show)
    ├───aarch64-linux
    │   ├───default omitted (use '--all-systems' to show)
    │   ├───notes omitted (use '--all-systems' to show)
    │   ├───todo omitted (use '--all-systems' to show)
    │   └───todo-done omitted (use '--all-systems' to show)
    ├───x86_64-darwin
    │   ├───default omitted (use '--all-systems' to show)
    │   ├───notes omitted (use '--all-systems' to show)
    │   ├───todo omitted (use '--all-systems' to show)
    │   └───todo-done omitted (use '--all-systems' to show)
    └───x86_64-linux
        ├───default: package 'notes'
        ├───notes: package 'notes'
        ├───todo: package 'todo'
        └───todo-done: package 'todo-done'
```

Now we can run the app running `nix run .#notes` or any of the other binaries built in this derivation with all its dependencies bundled with it.

## Conclusions

I use this notes app everyday, it serves me well, but I know I have a lot of opinions built into it. NeoVim, local files and bash, these are all choices that are aligned with a movement towards using linux tools I am into now. Building the "app" was pretty fun, only needed a few iterations and already landed on a pretty usable thing. Nix on the other hand was not easy to grasp, especially with the restrictions I made to the project. Documentation is not easily discoverable, `nix` as a build tool doesn't have good error messages and `nix` language being functional and declarative were some interesting challenges to me.

In the end I think it was worth it, now I can easily integrate it with my [NixOS configuration](https://github.com/cesarFuhr/cesarOS) and have it always baked in my system. As a next step in my Nix journey I would like to package a Go application with it, having all its dependencies versioned with Nix and a ready to use development shell.
