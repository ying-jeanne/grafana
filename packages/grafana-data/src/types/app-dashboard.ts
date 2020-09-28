import { Observable } from 'rxjs';
import { queryTemplates } from './queryTemplates';
import { datasourceSelector } from './components';

export function runDashboard(): Observable<Dashboard> {
    
    const dashboard = new Dashboard()

    dashboard.addTimePicker({from: '1h', to: 'now'});
    dashboard.addVariable(datasourceSelector);
        
    dashboard.addQuery(queryTemplates.PodLatency({
        ds: '$datasource',        
    });
    
    dashboard.runQueries().pipe(
        map((data: PanelData) => {
            
            dashboard.forEachSeries(data.series)
                .addPanel({ type: 'graph' })

        })        
    );

    return dashboard;
}