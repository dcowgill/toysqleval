# toysqleval

Toy SQL parser and evaluator, written to answer a technical homework problem.

Rough breakdown of time spent:

| Task                     | Hours |
| ------------------------ | -----:|
| lexical analyzer         |   2.0 |
| parser                   |   1.0 |
| basic evaluator          |   1.5 |
| comparisons, arithmetic  |   1.0 |
| input pos. in errors     |   1.0 |
| aggregate functions      |   2.0 |
| tests                    |   0.5 |
| **total**                |   9.0 |

Lines of code:

```
$ for d in `find * -type d -depth 0 | sort`; do
> cloc $d | perl -lne "next unless /^Go.*?(\d+)$/; print sprintf('%7s %5d', '$d', \$1)"
> done
    ast   250
    cmd    45
   eval  1075
  lexer   347
 parser   304
 pprint    61
  token   178
```
