import { useEffect, useState } from "react";
import type { PatternData } from "../util/types"

function usePatternDetails(id: number) {
    let [pattern, setPattern] = useState<PatternData | null>(null);
    useEffect(() => {
        //Fetch pattern info
        fetch(`http://localhost:3002/${id}`)
          .then((response) => response.json())
          .then((data) => setPattern(data))
          .catch((reason) => {
            console.log(reason);
            setPattern(null)
          });
    }, [id])
    return pattern;
}

function PatternCard({ id }: { id: number }) {
    const patternDetails = usePatternDetails(id);

    return <div>
        <img src="https://littleworldofwhimsy.com/wp-content/uploads/2023/09/IMG_5640.jpg" width={80} height={100} />
        <h2>{patternDetails?.name ?? "-"}</h2>
        <span>5/15</span>
    </div>
}

export default PatternCard