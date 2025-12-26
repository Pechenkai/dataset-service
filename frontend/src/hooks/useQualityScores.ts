import { useEffect, useMemo, useState } from 'react';
import { Dataset } from '../api/types';
import { qualityEngine } from '../wasm/qualityEngine';

export const useQualityScores = (datasets: Dataset[]) => {
  const [version, setVersion] = useState(0);

  useEffect(() => {
    qualityEngine.init().then(() => {
      setVersion(qualityEngine.getVersion());
    });
  }, []);

  const scores = useMemo(() => {
    const map: Record<number, number> = {};
    datasets.forEach((d) => {
      map[d.id] = qualityEngine.score(d);
    });
    return map;
  }, [datasets, version]);

  return { scores, wasmReady: qualityEngine.isReady() };
};
