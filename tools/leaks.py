#!/usr/bin/env python
# Open file leak checker
# Usage:
# strace -ff open,close,chdir,execve $REGULAR_CMD 2>&1 | python leaks.py
import re
import sys
import os

class ProcessHandler(object):
    def __init__(self, pid):
        self.pid = pid
        self.cwd = os.getcwd()
        self.handles = {}

    def on_open(self, path, flags, fd):
        if fd in self.handles:
            print('warning: pid %s fd %s already tracked (now: %s)' % (
                self.pid, fd, path))
        if path and path[0] != '/':
            path = self.cwd + path
        self.handles[fd] = (path, flags)

    def on_close(self, fd):
        if fd in self.handles:
            del self.handles[fd]

    def on_execve(self):
        for fd in self.handles.keys():
            path, flags = self.handles[fd]
            if 'O_CLOEXEC' in flags:
                self.on_close(fd)

    def on_chdir(self, path):
        self.cwd = path


class FileLeakChecker(object):
    PID_PATTERN = re.compile(r'^\[pid (\d+)\] ')
    OPEN_PATTERN = re.compile(r'^open\("([^"]*)", ([^,]+)[^\)]*\)\s*=\s*(\d+)$')
    CLOSE_PATTERN = re.compile(r'^close\((\d+)\)\s*=\s*(\d+)$')
    CHDIR_PATTERN = re.compile(r'^chdir\("([^"]*)"\)\s*=\s*(\d+)$')
    EXECVE_PATTERN = re.compile(r'^execve\(')

    line = ""

    def __init__(self):
        self.processes = {}

    def run(self, input):
        while True:
            self.line = input.readline()
            if not self.line:
                break
            self.process = self.get_process()
            self.strip_pid()
            if self.check_open():
                continue
            if self.check_close():
                continue
            if self.check_chdir():
                continue
            if self.check_execve():
                continue

    def get_process(self):
        pid = self.find_pid()
        if not pid in self.processes:
            self.processes[pid] = ProcessHandler(pid)
        return self.processes[pid]

    def find_pid(self):
        m = self.PID_PATTERN.match(self.line)
        if not m:
            return None
        return m.group(1)

    def check_open(self):
        m = self.OPEN_PATTERN.match(self.line)
        if m:
            self.process.on_open(m.group(1), m.group(2), m.group(3))
            return True

    def check_close(self):
        m = self.CLOSE_PATTERN.match(self.line)
        if m:
            self.process.on_close(m.group(1))
            return True

    def check_chdir(self):
        m = self.CHDIR_PATTERN.match(self.line)
        if m:
            self.process.on_chdir(m.group(1))
            return True

    def check_execve(self):
        m = self.EXECVE_PATTERN.match(self.line)
        if m:
            self.process.on_execve()
            return True

    def strip_pid(self):
        self.line = self.PID_PATTERN.sub("", self.line)

    def results(self):
        print("# Number of processes: %d" % len(self.processes))
        for process in self.processes.values():
            self.process_results(process)

    def process_results(self, process):
        if not process.handles:
            return
        for fd, data in process.handles.items():
            path, flags = data
            print("%s\t%s" % (process.pid, path))


if __name__ == '__main__':
    flc = FileLeakChecker()
    flc.run(sys.stdin)
    flc.results()
