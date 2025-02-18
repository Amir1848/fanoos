import type {RowSelectionOptions} from 'ag-grid-community';
import { HttpClient } from '@angular/common/http';
import { Component, EventEmitter, Input, OnInit, Output, ViewChild } from '@angular/core';
import { AgGridAngular } from 'ag-grid-angular';

@Component({
  selector: 'app-remote-data-grid',
  templateUrl: './remote-data-grid.component.html',
  styleUrl: './remote-data-grid.component.css'
})
export class RemoteDataGridComponent implements OnInit {
  @Input()
  dataviewKey: string = "";

  @Input()
  rowSelection: RowSelectionOptions<any> = {
    mode: 'multiRow',
    checkboxes: true,
  }

  @Output()
  selectionChanged: EventEmitter<any[]> = new EventEmitter<any[]>() 

  @Input() 
  colDefs:  any[] = [];

  @Input()
  parameters: any;

  @ViewChild("grd") 
  grid!: AgGridAngular;

  rowData: [] = [];

  constructor(
    private httpClient: HttpClient
  ){
    
  }

  refresh(){
    this.updateRemoteData()
  }

  updateRemoteData(){
    return this.httpClient.post<[]>(`/api/dataviews/${this.dataviewKey}`, {
      pageNumber: 1,
      pageSize: 1000,
      parameters: this.parameters
    }).subscribe(p => {
      this.rowData = p
    })
  }

  ngOnInit(): void {
   this.updateRemoteData()
  }

  gridSelectionChanged(p: any){
    this.selectionChanged.emit(p.api.getSelectedRows())
  }
}
