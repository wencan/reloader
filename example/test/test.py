#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# wencan
# 2019-02-01

import aiohttp
import asyncio


async def keep_call(url: str, loop: asyncio.AbstractEventLoop=None):
    if loop is None:
        loop = asyncio.get_event_loop()

    async with aiohttp.ClientSession(loop=loop) as session:
        while True:
            async with session.get(url) as response:
                while True:
                    line = await response.content.readline()
                    if line is b'':
                        break
                    print(line.decode("utf-8"), end='')

async def test(loop: asyncio.AbstractEventLoop):
    urls = ["http://127.0.0.1:809{}".format(i) for i in range(0, 10)]
    tasks = [keep_call(url, loop) for url in urls]
    await asyncio.wait(tasks, loop=loop)

if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.run_until_complete(test(loop))