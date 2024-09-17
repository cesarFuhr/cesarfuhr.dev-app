
##### September 14th, 2024

# What about Hare lang?
#### What about it? Its mascot is a [Hare](https://harelang.org/).

I always wanted to exercise my manual memory management muscles, but that is not a thing in my current job, where I use Go (backend) and Typescript (frontend). So, after hearing about it in [The Changelog](https://changelog.com/podcast/569) and [Developer Voices](https://www.youtube.com/watch?v=42y2Q9io3Xs), I got interested in Hare. A systems programming language with manual memory management, which aims to still be used 100 years from now.

You could ask me, why not just deep my toes in C? Well as much as I like C, I will be honest, I got hooked by the new and shiny. But could I write? At first I tried something big... What if I wrote a Hare Language Server (in Hare of course)? Although it felt exciting, I suddenly realised that I am not a systems programmer and I had so much to learn before tackling such a big task.

After dabbling with Hare implementing a few sorting algorithms, my next target was...

##### __Note:__ If you want to skip the article and just see the code, check it out [here](https://git.sr.ht/~cesarfuhr/common-sense-guide-dt-algos/tree/main/item/hashmap.ha).

## A Hashmap in Hare

One of the most basic data structures you can have, the hashmap is usually already available as a language feature. In Go we have `map[keyType]valueType`, a generic data structure that, in the average case, can find any key in O(n) time complexity. In Javascript the base object type is basically a hashmap, where the object fields are the keys and their values the hashmap values.

Well, in Hare land, there is no hashmap ready to go, if you need one you need to code it yourself. Not even in the standard library? Nope.


