
##### May 15th, 2024

# Packaging Go with Nix
#### I wonder what Go application should I wrap with Nix... 

In my [last post](https://cesarfuhr.dev/blog/packaging_bash_wth_nix.html) I packaged a small set of bash scripts, which I use almost every day as my notes taking app. I had a few objectives writing the Nix code to package it: use only the Nix "standard library" and use Nix flakes. That was fun, but quite challenging given my background with imperative languages (such as Go and C), while Nix is declarative. Also the experience I have with packaging an application is pretty much reduced to using `docker`, which, although pretty effective and common, does not share with Nix the reproducibility nor hermetic build process. So, although I have some familiarity with Nix by using NixOS, packaging stuff with it is a completely new area I am exploring. 

This time lets go a little further, lets try an actual service like this blog. Perfect! 

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

Does not look like rocket science, does it?
