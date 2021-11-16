import * as d3 from 'd3';

// credit: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Math/random
function getRandomIntInclusive(minimum, maximum) {
  const min = Math.ceil(minimum);
  const max = Math.floor(maximum);
  // The maximum is inclusive and the minimum is inclusive
  return Math.floor(Math.random() * (max - min + 1) + min);
}

function addElement(nodes, links) {
  const randomInt = getRandomIntInclusive(1, 100);
  if (randomInt < 3 && nodes.length > 10) {
    // add single Node

    nodes.push({ id: nodes.length + 1 });
  } else if (randomInt < 6 && nodes.length > 10) {
    // add link
    links.push({
      source: getRandomIntInclusive(1, nodes.length - 1),
      target: getRandomIntInclusive(1, nodes.length),
    });
  } else {
    // add attached Node
    const index = nodes.length + 1;
    const newNode = { id: index };
    const newLink = { source: getRandomIntInclusive(1, index - 1), target: index };
    links.push(newLink);
    nodes.push(newNode);

    // set coordinates of new node to the coordinates of its parent
    const parentNode = nodes[newLink.source - 1];
    newNode.x = parentNode.x;
    newNode.y = parentNode.y;
  }
}

// RandomGraph adds random connected nodes to the svg
export default class RandomGraph {
  constructor(width, height, svgId) {
    this.svgId = svgId;
    this.width = width;
    this.height = height;

    // add attributes to root svg
    this.rootSvg = d3.select(`#${this.svgId}`).attr('viewBox', `0 0 ${this.width} ${this.height}`);
    this.rootGroup = this.rootSvg.append('g').attr('class', 'root-group');

    this.lineGroup = this.rootGroup.append('g')
      .attr('stroke', '#999')
      .attr('stroke-opacity', 0.6);

    this.nodeGroup = this.rootGroup
      .append('g')
      .attr('stroke', '#fff')
      .attr('stroke-width', '#C2C2C2');

    this.simulation = null;
    this.animationTimer = null;
  }

  // startDrawLoop starts an interval timer which adds nodes at random position in the graph.
  // Call stopDrawLoop to stop the timer.
  startDrawLoop(intervalMs) {
    const nodes = [{ id: 1 }];
    const links = [];

    this.animationTimer = setInterval(() => {
      const iterations = getRandomIntInclusive(1, 10);

      for (let i = 0; i < iterations; i += 1) {
        addElement(nodes, links);
      }

      this.stopSimulation();
      this.draw(nodes, links);
      if (nodes.length > 150) {
        clearInterval(this.animationTimer);
      }
    }, intervalMs);
  }

  stopDrawLoop() {
    if (this.animationTimer !== null) {
      clearInterval(this.animationTimer);
    }
  }

  draw(nodes, links) {
    this.simulation = d3.forceSimulation(nodes)
      .alphaDecay(0.08)
      .force('link', d3.forceLink(links).id((d) => d.id))
      .force('charge', d3.forceManyBody().strength(-80))
      .force('x', d3.forceX(this.width / 2).strength(0.04))
      .force('y', d3.forceY(this.height / 2).strength(0.04))
      .force('center', d3.forceCenter(this.width / 2, this.height / 2));

    const link = this.lineGroup
      .selectAll('.arrow')
      .data(links, (d) => `${d.source}${d.target}`)
      .join('line')
      .attr('class', 'arrow')
      .attr('stroke-width', 1);

    const node = this.nodeGroup
      .selectAll('circle')
      .data(nodes, (d) => d.id)
      .join('circle')
      .attr('r', 7)
      .attr('fill', '#008ee5');

    this.simulation.on('tick', () => {
      link
        .attr('x1', (d) => d.source.x)
        .attr('y1', (d) => d.source.y)
        .attr('x2', (d) => d.target.x)
        .attr('y2', (d) => d.target.y);

      node
        .attr('cx', (d) => d.x)
        .attr('cy', (d) => d.y);
    });
  }

  stopSimulation() {
    if (this.simulation !== null) this.simulation.stop();
  }
}
