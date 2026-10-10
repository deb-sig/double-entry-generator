// Package xls reads BIFF5/BIFF8 (.xls) workbooks.
//
// Copied from github.com/extrame/xls (Apache-2.0, see LICENSE in this
// directory) at v0.0.2-0.20200426124601-4a6cf263071b. Modified by the
// double-entry-generator authors: NumberCol no longer formats every
// non-General cell as a date (built-in number formats used to panic on a
// missing Formats entry, and "#,##0.00" amounts came out as dates); the
// xlsx comparison helper and its tealeg/xlsx dependency are dropped.
package xls
