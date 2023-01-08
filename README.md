# respool

[![golang](https://img.shields.io/badge/Language-Go-green.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/github.com/mwat56/respool?status.svg)](https://godoc.org/github.com/mwat56/respool)
[![Go Report](https://goreportcard.com/badge/github.com/mwat56/respool)](https://goreportcard.com/report/github.com/mwat56/respool)
[![Issues](https://img.shields.io/github/issues/mwat56/respool.svg)](https://github.com/mwat56/respool/issues?q=is%3Aopen+is%3Aissue)
[![Size](https://img.shields.io/github/repo-size/mwat56/respool.svg)](https://github.com/mwat56/respool/)
[![Tag](https://img.shields.io/github/tag/mwat56/respool.svg)](https://github.com/mwat56/respool/tags)
[![View examples](https://img.shields.io/badge/learn%20by-examples-0077b3.svg)](https://github.com/mwat56/respool/blob/main/_demo/demo.go)
[![License](https://img.shields.io/github/mwat56/respool.svg)](https://github.com/mwat56/respool/blob/main/LICENSE)

----
<!-- TOC -->

- [respool](#respool)
	- [Purpose](#purpose)
	- [Installation](#installation)
	- [Usage](#usage)
	- [Libraries](#libraries)
	- [Credits](#credits)
	- [Licence](#licence)

<!-- /TOC -->
## Purpose

Once in a while you'll find yourself needing some kind of a pool that holds some recyclable resources like – for example – database connections (which can be quite expensive).
A FiFo like (First In First Out) list comes to mind which could take any resources and give them back.
However, `Go` provides a much smarter data structure for this purpose: `channels`. And such a `channel` serves internally as the backbone of this resources pool.

## Installation

You can use `Go` to install this package for you:

    go get -u github.com/mwat56/respool

## Usage

    //TODO

## Libraries

No external libraries were used building `respool`.

## Credits

This package is based on the "pool" example in

	William Kennedy, Brian Ketelsen, Erik St. Martin:
	Go in Action; Shelter Island: Manning, 2015; Chapter 7
	Example provided with help from Fatih Arslan and Gabriel Aszalos.

with some modifications and additions from me.

## Licence

        Copyright © 2023 M.Watermann, 10247 Berlin, Germany
                        All rights reserved
                    EMail : <support@mwat.de>

> This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
>
> This software is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
>
> You should have received a copy of the GNU General Public License along with this program. If not, see the [GNU General Public License](http://www.gnu.org/licenses/gpl.html) for details.

----
[![GFDL](https://www.gnu.org/graphics/gfdl-logo-tiny.png)](http://www.gnu.org/copyleft/fdl.html)
