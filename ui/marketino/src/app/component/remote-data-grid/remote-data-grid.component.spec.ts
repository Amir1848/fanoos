import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RemoteDataGridComponent } from './remote-data-grid.component';

describe('RemoteDataGridComponent', () => {
  let component: RemoteDataGridComponent;
  let fixture: ComponentFixture<RemoteDataGridComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [RemoteDataGridComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(RemoteDataGridComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
