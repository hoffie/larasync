// build windows

package repository

import (
    "github.com/hoffie/larasync/helpers/path"

    . "gopkg.in/check.v1"
)


// Windows should hide the newly created management directories.
func (t *RepositoryTests) TestManagementDirMarkAsHidden(c *C) {
    r := New(t.dir)
    err := r.CreateManagementDir()
    c.Assert(err, IsNil)

    check, err := path.IsHidden(r.GetManagementDir())
    c.Assert(err, IsNil)
    c.Assert(check, Equals, true)
}
