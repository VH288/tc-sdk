package pagination_test

import (
	"testing"

	"github.com/VH288/tc-sdk/pagination"
)

func TestPaginationBoundaries(t *testing.T) {
	for _, tc := range []struct{ page, perPage, wantPage, wantPerPage, wantOffset int }{
		{0, 0, 1, 10, 0},
		{2, 101, 2, 100, 100},
	} {
		page, perPage, offset := pagination.NormalizePagination(tc.page, tc.perPage)
		if page != tc.wantPage || perPage != tc.wantPerPage || offset != tc.wantOffset {
			t.Errorf("NormalizePagination(%d, %d) = (%d, %d, %d)", tc.page, tc.perPage, page, perPage, offset)
		}
	}
	if got := pagination.TotalPages(21, 10); got != 3 {
		t.Errorf("TotalPages(21, 10) = %d", got)
	}
	if got := pagination.TotalPages(21, 0); got != 0 {
		t.Errorf("TotalPages(21, 0) = %d", got)
	}
}
