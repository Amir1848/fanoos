import { CommonModule } from '@angular/common';
import {
  ChangeDetectionStrategy,
  Component,
  computed,
  effect,
  input,
  output,
  signal,
} from '@angular/core';
import { IconComponent } from '../icon/icon.component';

export interface ColumnDef {
  key: string;
  header: string;
  width?: string;
  /** Force direction for latin values (emails, usernames, numbers). */
  ltr?: boolean;
  type?: 'text' | 'boolean';
  /** Labels for the boolean badge, defaults to بله / خیر */
  booleanLabels?: [string, string];
}

export interface TableRow {
  id: number;
}

export const PAGE_SIZES = [10, 25, 50] as const;

@Component({
  selector: 'app-data-table',
  standalone: true,
  imports: [CommonModule, IconComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './data-table.component.html',
  styleUrl: './data-table.component.scss',
})
export class DataTableComponent {
  readonly columns = input.required<ColumnDef[]>();
  readonly rows = input.required<TableRow[]>();
  readonly loading = input(false);
  readonly page = input(1);
  readonly pageSize = input(10);
  /** true when the current page is full, so a next page may exist */
  readonly hasNext = input(false);

  readonly pageChange = output<number>();
  readonly pageSizeChange = output<number>();
  readonly rowOpen = output<TableRow>();
  readonly selectionChange = output<number[]>();

  readonly pageSizes = PAGE_SIZES;
  readonly selected = signal<ReadonlySet<number>>(new Set());

  readonly allSelected = computed(() => {
    const rows = this.rows();
    const selected = this.selected();
    return rows.length > 0 && rows.every((r) => selected.has(r.id));
  });

  constructor() {
    // Rows changed (new page / refresh): drop the selection.
    effect(() => {
      this.rows();
      this.selected.set(new Set());
      this.selectionChange.emit([]);
    });
  }

  toggleRow(row: TableRow, checked: boolean): void {
    const next = new Set(this.selected());
    if (checked) {
      next.add(row.id);
    } else {
      next.delete(row.id);
    }
    this.selected.set(next);
    this.selectionChange.emit([...next]);
  }

  toggleAll(checked: boolean): void {
    const next = checked ? new Set(this.rows().map((r) => r.id)) : new Set<number>();
    this.selected.set(next);
    this.selectionChange.emit([...next]);
  }

  cellValue(row: TableRow, col: ColumnDef): unknown {
    return (row as unknown as Record<string, unknown>)[col.key];
  }

  booleanLabel(value: unknown, col: ColumnDef): string {
    const [yes, no] = col.booleanLabels ?? ['بله', 'خیر'];
    return value ? yes : no;
  }
}
