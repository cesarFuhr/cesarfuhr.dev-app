
##### May 20th, 2024

# Packaging Go with Nix
#### I wonder what Go application should I wrap with Nix... 

In my [last post](https://cesarfuhr.dev/blog/packaging_bash_wth_nix.html) I packaged a small set of bash scripts, which I use almost every day as my notes taking app. I had a few objectives writing the Nix code to package it: use only the Nix "standard library" and use Nix flakes. That was fun, but quite challenging given my background with imperative languages (such as Go and C), while Nix is declarative. Also the experience I have with packaging an application is pretty much reduced to using `docker`, which, although pretty effective and common, does not share with Nix the reproducibility nor hermetic build process. So, although I have some familiarity with Nix by using NixOS, packaging stuff with it is a completely new area I am exploring. 

This time lets go a little further, lets try an actual service like this blog. Perfect! 

##### __Note:__ If you want to skip the article and just see the code, check it out [here](https://github.com/cesarFuhr/cesarFuhr.dev-app).

## How's this blog implemented?

I started [this blog](https://github.com/cesarFuhr/cesarfuhr.dev-app) with a simple http file server, but writing every single page in `HTML`. That quickly got me tired of writing posts and my motivation to write posts dropped severely. I kept delaying writing more posts although I had good topics to write about, because of this DX issue. Eventually I decided to rewrite the blog, make it so I could write Markdown for the posts and generate the HTML based on it. I scrapped Github for an Go library that converted Markdown to HTML and eventually found a good candidate. 

After some hard work porting the old HTML headers and footers into Go templates and rewriting the posts in Markdown I landed on the new implementation, it used:

- A Markdown to HTML [lib](github.com/gomarkdown/markdown).
- Code generation to put the files together.
- Go embed to create a in memory file server. 

And that is it. I wanted something tiny and easy to maintain, after all it is just a little blog that serves some HTML files. I kept everything else the same, where I hosted it ([fly.io](https://fly.io)) and a Docker container as the deployment artifact.

## Go and Nix

Being so simple, but still requiring external libraries and having multiple build steps (generation and compilation), my little personal blog looked like a good candidate as my next Nix challenge.

I kept the same restrictions as the time I packaged the [bash scripts](http://localhost:8080/blog/packaging_bash_wth_nix.html), no Nix external libs and use flakes. After some research I found a few posts talking about Go and Nix, but every single one was using [flake-utils](https://github.com/numtide/flake-utils) or [flake-parts](https://flake.parts/) and, I get it, without them you are required to write a good portion of boilerplate code, but since learning was my primary concern I didn't followed them directly.

The first thing I needed to know is how to build the Go code with its dependencies. To do that Nix packages already have a solution, based on the `go.mod` and `go.sum` file, `p.buildGoModule` build Go programs in two phases: first it fetches the all the external modules the app needs and "vendors" them in a intermediate derivation, after that it uses this intermediate derivation results to build the final output.

```nix
blog = nixpkgs.buildGoModule
  {
    # Binary name.
    name = name;
    # In the first run you will might want set vendorHash to lib.fakeHash.
    vendorHash = "sha256-K6hdGsOjCJLx1nH69MHoTzV9tD05Gz4LdGGccCL1TOk=";
    src = ./.;
    # This specifies which package to build, otherwise
    # all the packages will be built.
    subPackages = [ "cmd/blog" ];

    # Add any environment variables you want in build time.
    CGO_ENABLED = 0;

    # This runs before the build step, I have to generate the
    # HTML based on the Markdown.
    # This is done by the `pre` make target.
    preBuild = ''
      make pre
    '';
  };
```

Not rocket science, right? 

This will create a derivation with the blog binary in it, ready to be ran. So lets do that, we would still need some boilerplate code to make it an actual Nix packages, but lets not worry about that now. Running `nix run .#blog` runs the blog exposing it on `http://localhost:8080`.

The next step is put it inside a container so I could deploy it to `fly`. Nix has a built in function to do that too, `p.dockerTools.buildImage` will receive a (almost one to one) configuration and create a minimal image, that only contains what the blog needs to run.

```nix
container = p.dockerTools.buildImage {
  name = name;
  tag = "latest";
  config = {
    # `blog` in this string interpolation expands to the derivation path.
    # /bin/${name} is where the binary will be after the build.
    Cmd = [ "${blog}/bin/${name}" ];
  };
};
```

Running `nix build .#container` will build the image and set `result` (the file in the flake root directory) as the image tarball. You can load the image into your container runtime (if its docker) with: `docker load < result`. This takes care of the application and the deployment artifact, but if you would like to develop such application locally you would need: `go` and some `go-tools`, `make` and the `fly` CLI. `nix develop` is the Nix tool to create fully pinned and reproducible development shells, so lets declare not only how to build the Go binary but also whats needed to do that in the `devShells` attribute.

```nix
devShells = systemsToAttrs
  (system:
    let
      p = import nixpkgs { system = system; };
    in
    {
      default =
        p.mkShell {
          # This is where you list your build dependencies.
          # They will be available in your `nix develop` shell also.
          buildInputs = [
            p.flyctl
            p.go
            p.go-tools
            p.gopls
            p.gnumake
          ];
        };
    })
  systems;
```

## Some final touches

I mentioned there was some boilerplate code missing, that would wrap all this things we wrote. For instance, the code to build the blog for each supported system. If you got this far into the post, maybe you are interested in that.

To clean the code and make the important things more recognizable I extracted most of the systems logic to a function: `forEachSystem`. This function takes a callback function and calls it for each system received as the second argument, passing the system specific `nixpkgs`. That enabled me to create system specific versions of the packages for each of the derivations final output, without having to repeat the system loops. In the end I had the following.

```nix
{
  description = "cesarfuhr.dev simple blog";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs = inputs@{ nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forEachSystem = (callback: builtins.listToAttrs (
        builtins.map
          (system:
            let
              pkgs = import nixpkgs { system = system; };
            in
            {
              name = system;
              value = callback pkgs;
            })
          systems
      )
      );
    in
    {
      devShells = forEachSystem
        (pkgs: {
          default = { /* ... */ };
        });

      packages = forEachSystem
        (pkgs:
          let
            name = "blog";
          in
          rec {
            default = blog;
            blog = pkgs.buildGoModule { /* ... */ };
            container = pkgs.dockerTools.buildImage { /* ... */ };
          });
    };
}
```

Well, as a second attempt of packaging something with Nix that wasn't bad at all. This time having some code to base the new implementation I felt it was way more approachable, to the point that I could even push a little further the refactoring of the system specific logic.

I am getting the hand of it, I don't know. What I know is am enjoying this learning process, I wonder what would be the next thing I can point my Nix cannon to...
